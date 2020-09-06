package gw2util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
)

// Gw2Api data used to access the json api.
type Gw2Api struct {
	BaseURL string
	Key     string
}

// UserData contains the connection between the chat name and an
// gw2 account char name.
type UserData struct {
	GameID   string
	ChatName string
	Key      string
}

// UserDataSlice contains all users data
type UserDataSlice []UserData

// World is the gw2 world json in go
type World struct {
	ID     int64
	Name   string
	Colour string
}

// WorldNames all worlds in this matchup
type WorldNames struct {
	//worldid worldname
	WorldName map[int64]string
}

// WorldColours the colurs of our worlds (red,blue,green)
type WorldColours struct {
	//worldid worldname
	WorldColour map[int64]string
}

// CharacterCache since some data takes long time to parse and fetch we cache it
type CharacterCache struct {
	// need mutex here if called from goroutine..
	charactersCache map[string]*gabs.Container
	age             map[string]int64
}

// AccountCache more caching of not so often uptaded data
type AccountCache struct {
	accountCache map[string]*gabs.Container
	age          map[string]int64
}

// WWWScore contains the current score in matchup
type WWWScore struct {
	Red   int64
	Blue  int64
	Green int64
}

var accountCache = AccountCache{accountCache: make(map[string]*gabs.Container), age: make(map[string]int64)}
var charCache = CharacterCache{charactersCache: make(map[string]*gabs.Container), age: make(map[string]int64)}

// Update a users gw2 api key data or add the user to the slice if
// the user hasnt set his data yet.
func (b UserDataSlice) Update(user UserData) UserDataSlice {
	found := false
	for _, item := range b {
		if item.ChatName == user.ChatName {
			item = user
			found = true
		}
	}
	if !found {
		b = append(b, user)
	}

	return b
}

func getCrafting(chars *gabs.Container, name string) []GW2Crafting {
	size, _ := chars.ArrayCount("name")
	//fmt.Println(name)
	var retVal = make([]GW2Crafting, 0)

	for index := 0; index < size; index++ {
		if strings.Contains(chars.Index(index).Search("name").String(), name) {
			crafts := chars.Index(index).Search("crafting")
			discipline := crafts.Search("discipline")
			rating := crafts.Search("rating")
			active := crafts.Search("active")
			amountD, _ := crafts.ArrayCount("discipline")
			amountR, _ := crafts.ArrayCount("rating")
			amountA, _ := crafts.ArrayCount("active")

			if (amountD != amountR) && (amountD != amountA) {
				return retVal
			}
			for n := 0; n < amountD; n++ {
				Discipline := discipline.Index(n).String()
				Rating, _ := strconv.ParseInt(rating.Index(n).String(), 10, 64)
				Active, _ := strconv.ParseBool(active.Index(n).String())
				retVal = append(retVal, GW2Crafting{Discipline, Rating, Active})
			}
		}
	}

	return retVal
}

func getCacheCharactersStruct(gw2 Gw2Api) *gabs.Container {
	var jsonParsed *gabs.Container

	value, ok := charCache.age[gw2.Key]
	if (value+30 > time.Now().Unix()) && ok {
		fmt.Println("using cached values")
		jsonParsed = charCache.charactersCache[gw2.Key]
	} else {
		fmt.Println("getting new values")
		body, _ := QueryAnetAuth(gw2, "characters")
		jsonParsed, _ = gabs.ParseJSON(body)
		charCache.charactersCache[gw2.Key] = jsonParsed
		charCache.age[gw2.Key] = time.Now().Unix()
	}

	return jsonParsed
}

// GetCrafting returns all crafts for all charachters
func GetCrafting(gw2 Gw2Api, name string) []GW2Crafting {
	jsonParsed := getCacheCharactersStruct(gw2)

	return getCrafting(jsonParsed, name)
}

// FindItem search for items in a characters bags
func FindItem(gw2 Gw2Api, charName string, item string) []*GW2Item {

	jsonParsed := getCacheCharactersStruct(gw2)
	//fmt.Println(jsonParsed.StringIndent("", "\t"))
	items := GetItems(gw2, GetItemIdsFromBags(jsonParsed, charName))
	return findItem(items, item)
}

// Tries to find a guild wars 2 item from item name or item detail.type
func findItem(itemArr *gabs.Container, itemName string) []*GW2Item {
	var retVal []*GW2Item
	items := itemArr.Children()

	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Search("name").String()), strings.ToLower(itemName)) ||
			strings.Contains(strings.ToLower(item.Path("details.type").String()), strings.ToLower(itemName)) {
			//fmt.Println(item.StringIndent("", "\t"))
			var gwItem GW2Item
			dec := json.NewDecoder(strings.NewReader(item.String()))
			dec.UseNumber()
			err := dec.Decode(&gwItem)
			if err != nil {
				fmt.Println("Error during conversion ", err)
			} else {
				retVal = append(retVal, &gwItem)
			}
		}
	}
	return retVal
}

// GetItems query the guild wars 2 json api for the items thsi character has in its bags.
func GetItems(gw2 Gw2Api, Ids []uint64) *gabs.Container {
	strIds := arrayToString(Ids, ",")
	body, _ := QueryAnet(gw2, "items", "ids", strIds)
	jsonParsed, _ := gabs.ParseJSON(body)
	return jsonParsed
}

// GetItemIdsFromBags extract all items from the json blob
func GetItemIdsFromBags(chars *gabs.Container, charName string) []uint64 {
	var retVal []uint64
	//fmt.Println(charName)
	children := chars.Children()
	for _, char := range children {
		if strings.Contains(strings.ToLower(char.S("name").String()), strings.ToLower(charName)) {
			items := char.Path("bags.inventory.id")
			//fmt.Println(items)
			retVal = append(retVal, flattenIDArray(items)...)
			//fmt.Println((retVal))
		}
	}
	return retVal
}

func getCharacterNames(chars *gabs.Container) []string {
	var retVal []string
	children := chars.Search("name").Children()

	for _, char := range children {
		retVal = append(retVal, strings.Trim(char.String(), "\""))
	}
	return retVal
}

// GetCharacterNames Get all character names from an account
func GetCharacterNames(gw2 Gw2Api) []string {
	body, _ := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	return getCharacterNames(jsonParsed)
}

// ReadUserData read in account info and api keys from disc
func ReadUserData(filename string) UserDataSlice {
	var retVal UserDataSlice

	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
		return nil
	}
	err = json.Unmarshal(buff, &retVal)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	return retVal
}

// SaveUserData save username and api key to disc
func SaveUserData(userData UserDataSlice) string {

	userJSON, err := json.Marshal(userData)
	if err != nil {
		fmt.Print(err)
		return "Can't marshall data"
	}
	err = ioutil.WriteFile("data.dat", userJSON, 0600) // just pass the file name
	if err != nil {
		fmt.Print(err)
		return "Can't write to file"
	}
	return ""
}

// GetUserData return a user account from a discord chat name
func GetUserData(users UserDataSlice, chatName string) UserData {
	//for _, tuser := range users {
	//    fmt.Printf("username = %s GameId = %s Key = %s\n", tuser.ChatName, tuser.GameID, tuser.Key)
	//}
	for _, user := range users {
		fmt.Println(user)
		if user.ChatName == chatName {
			return user
		}
	}

	return UserData{"", "", ""}
}

// UpsertUserData update or insert gw2 api user info for a discord chat name
func UpsertUserData(users UserDataSlice, user UserData) UserDataSlice {
	found := false
	for index, item := range users {
		if item.ChatName == user.ChatName {
			users[index] = user
			found = true
			fmt.Println("found")
		}
	}
	if !found {
		users = append(users, user)
	}

	return users
}

func getCacheAccountStruct(gw2 Gw2Api) *gabs.Container {
	var jsonParsed *gabs.Container

	value, ok := accountCache.age[gw2.Key]
	if (value+300 > time.Now().Unix()) && ok {
		fmt.Println("using cached values")
		jsonParsed = accountCache.accountCache[gw2.Key]
	} else {
		fmt.Println("getting new values")
		body, _ := QueryAnetAuth(gw2, "account")
		jsonParsed, _ = gabs.ParseJSON(body)
		accountCache.accountCache[gw2.Key] = jsonParsed
		accountCache.age[gw2.Key] = time.Now().Unix()
	}

	return jsonParsed
}

// GetHomeWorld for a gw2 account
func GetHomeWorld(gw2 Gw2Api) string {
	jsonParsed := getCacheAccountStruct(gw2)
	return jsonParsed.Search("world").String()
}

func extractWvWvWStats(child *gabs.Container) GW2WvWvWStats {

	retVal := GW2WvWvWStats{}

	retVal.ID = child.S("id").String()
	value, ok := child.S("type").Data().(string)
	if ok {
		retVal.Name = strings.ToLower(strings.Replace(value, "Home", "", -1))
	} else {
		retVal.Name = ""
	}
	retVal.Deaths.Red, ok = child.Path("deaths.red").Data().(float64)
	retVal.Deaths.Blue, ok = child.Path("deaths.blue").Data().(float64)
	retVal.Deaths.Green, ok = child.Path("deaths.green").Data().(float64)

	retVal.Kills.Red, ok = child.Path("kills.red").Data().(float64)
	retVal.Kills.Blue, ok = child.Path("kills.blue").Data().(float64)
	retVal.Kills.Green, ok = child.Path("kills.green").Data().(float64)

	if !ok {
		log.Fatal("No data to parse. ")
		log.Fatal(child)
	}
	return retVal
}

// GetWWWStats returns the current WvWvW standings
func GetWWWStats(gw2 Gw2Api, world string) [5]GW2WvWvWStats {
	retVal := [5]GW2WvWvWStats{}
	item := GW2WvWvWStats{}
	var body []byte
	var err error
	for retries := 0; retries < 5; retries++ {
		body, err = QueryAnet(gw2, "wvw/matches/stats", "world", world)
		if err == nil {
			break
		}
		log.Printf("No data back retry number %d\n", retries+1)
		time.Sleep(100 * time.Millisecond)
	}
	jsonParsed, err := gabs.ParseJSON(body)

	if err != nil {
		log.Fatal("error code:", err, "\n", jsonParsed.String())
	}
	fmt.Println(jsonParsed.StringIndent("", "  "))
	//fmt.Printf("key: %v,value: %v\n", 0, jsonParsed)
	item = extractWvWvWStats(jsonParsed)
	item.Name = "Total"

	retVal[0] = item

	values := jsonParsed.S("maps").Children()

	if err != nil {
		log.Fatal("error code:", err, "\n", jsonParsed.String())
	}
	fmt.Println(len(values))
	for key, child := range values {
		//fmt.Printf("key: %v,value: %v\n", key+1, child)
		retVal[key+1] = extractWvWvWStats(child)
	}

	return retVal
}

// GetWvWvWScore Get current score in www
func GetWvWvWScore(gw2 Gw2Api, world string) WWWScore {
	var retVal WWWScore
	var body []byte
	var err error
	for retries := 0; retries < 5; retries++ {
		body, err = QueryAnet(gw2, "wvw/matches/scores", "world", world)
		if err == nil {
			break
		}
		log.Printf("No data back retry number %d\n", retries+1)
		time.Sleep(100 * time.Millisecond)
	}
	jsonParsed, err := gabs.ParseJSON(body)

	if err != nil {
		log.Fatal("error code:", err, "\n", jsonParsed.String())
	}
	//values := jsonParsed.S("scores").Children()

	//fmt.Println(values)
	retVal.Red = int64(jsonParsed.Path("scores.red").Data().(float64))
	retVal.Blue = int64(jsonParsed.Path("scores.blue").Data().(float64))
	retVal.Green = int64(jsonParsed.Path("scores.green").Data().(float64))
	return retVal
}

// GetWorlds returns the names of the servers in WvWvW
func GetWorlds(gw2 Gw2Api, worlds string) WorldNames {
	retVal := WorldNames{WorldName: make(map[int64]string)}

	body, _ := QueryAnet(gw2, "worlds", "ids", worlds)
	dec := json.NewDecoder(strings.NewReader(string(body[:])))
	dec.UseNumber()
	jsonParsed, _ := gabs.ParseJSONDecoder(dec)

	items := jsonParsed.Children()
	for _, item := range items {
		key, _ := item.Path("id").Data().(json.Number).Int64()
		retVal.WorldName[key] = item.Path("name").Data().(string)
	}

	return retVal
}

// GetWvWvWColours return main worlds from this matchup
func GetWvWvWColours(gw2 Gw2Api, worldID string) WorldColours {
	retVal := WorldColours{WorldColour: make(map[int64]string)}

	body, _ := QueryAnet(gw2, "wvw/matches", "world", worldID)
	dec := json.NewDecoder(strings.NewReader(string(body[:])))
	dec.UseNumber()
	jsonParsed, _ := gabs.ParseJSONDecoder(dec)

	items := jsonParsed.S("worlds").ChildrenMap()
	for colour, id := range items {
		//key, _ := item.Path("id").Data().(json.Number).Int64()
		//fmt.Printf("key=%v value=%v\n", id, colour)
		tmpID, _ := id.Data().(json.Number).Int64()
		retVal.WorldColour[tmpID] = colour
	}
	return retVal
}
