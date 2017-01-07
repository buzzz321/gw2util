package gw2util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
)

type Gw2Api struct {
	BaseUrl string
	Key     string
}

type UserData struct {
	GameId   string
	ChatName string
	Key      string
}

type UserDataSlice []UserData

type CharacterCache struct {
	// need mutex here if called from goroutine..
	charactersCache map[string]*gabs.Container
	age             map[string]int64
}

var cache = CharacterCache{charactersCache: make(map[string]*gabs.Container), age: make(map[string]int64)}

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

func GetCrafting(gw2 Gw2Api, name string) []GW2Crafting {
	var jsonParsed *gabs.Container

	value, ok := cache.age[name]
	if (value+30 > time.Now().Unix()) && ok {
		fmt.Println("using cached values")
		jsonParsed = cache.charactersCache[name]
	} else {
		fmt.Println("getting new values")
		body := QueryAnetAuth(gw2, "characters")
		jsonParsed, _ = gabs.ParseJSON(body)
		cache.charactersCache[name] = jsonParsed
		cache.age[name] = time.Now().Unix()
	}
	return getCrafting(jsonParsed, name)
}

// Tries to find a guild wars 2 item from item name or item detail.type
func FindItem(itemArr *gabs.Container, itemName string) []*GW2Item {
	var retVal []*GW2Item
	items, _ := itemArr.Children()

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

// Query the guild wars 2 json api for the items thsi character has in its bags.
func GetItems(gw2 Gw2Api, Ids []uint64) *gabs.Container {
	strIds := arrayToString(Ids, ",")
	body := QueryAnet(gw2, "items", "ids", strIds)
	jsonParsed, _ := gabs.ParseJSON(body)
	return jsonParsed
}

// Extract all items from the json blob
func GetItemIdsFromBags(chars *gabs.Container, charName string) []uint64 {
	var retVal []uint64
	children, _ := chars.Children()
	for _, char := range children {
		if strings.Contains(strings.ToLower(char.S("name").String()), charName) {
			items := char.Path("bags.inventory.id")
			//fmt.Println(items)
			retVal = append(retVal, flattenIdArray(items)...)
			//fmt.Println((retVal))
		}
	}
	return retVal
}

func getCharacterNames(chars *gabs.Container) []string {
	var retVal []string
	children, _ := chars.Search("name").Children()

	for _, char := range children {
		retVal = append(retVal, strings.Trim(char.String(), "\""))
	}
	return retVal
}

func GetCharacterNames(gw2 Gw2Api) []string {
	body := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	return getCharacterNames(jsonParsed)
}

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

func SaveUserData(userData UserDataSlice) string {

	userJson, err := json.Marshal(userData)
	if err != nil {
		fmt.Print(err)
		return "Can't marshall data"
	}
	err = ioutil.WriteFile("data.dat", userJson, 0600) // just pass the file name
	if err != nil {
		fmt.Print(err)
		return "Can't write to file"
	}
	return ""
}

func GetUserData(users UserDataSlice, chatName string) UserData {
	for _, user := range users {
		if user.ChatName == chatName {
			return user
		}
	}

	return UserData{"", "", ""}
}

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

/*
func main() {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	body := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	getCharacterNames(jsonParsed)
	//fmt.Println(jsonParsed.StringIndent("","\t"))
	craftings := getCrafting(jsonParsed, "Notamik")
	log.Println(craftings[0])
	items := GetItems(gw2, GetItemIdsFromBags(jsonParsed, "nomitik"))
	fmt.Println(FindItem(items, "food"))
	return
}
*/
