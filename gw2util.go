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

type UserDataSlice []UserData

type Worlds struct {
    //worldid worldname
    WorldName map[int64]string
}

type CharacterCache struct {
    // need mutex here if called from goroutine..
    charactersCache map[string]*gabs.Container
    age             map[string]int64
}

type AccountCache struct {
    accountCache map[string]*gabs.Container
    age          map[string]int64
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
        body := QueryAnetAuth(gw2, "characters")
        jsonParsed, _ = gabs.ParseJSON(body)
        charCache.charactersCache[gw2.Key] = jsonParsed
        charCache.age[gw2.Key] = time.Now().Unix()
    }

    return jsonParsed
}

func GetCrafting(gw2 Gw2Api, name string) []GW2Crafting {
    jsonParsed := getCacheCharactersStruct(gw2)

    return getCrafting(jsonParsed, name)
}

func FindItem(gw2 Gw2Api, charName string, item string) []*GW2Item {

    jsonParsed := getCacheCharactersStruct(gw2)
    items := GetItems(gw2, GetItemIdsFromBags(jsonParsed, charName))
    return findItem(items, item)
}

// Tries to find a guild wars 2 item from item name or item detail.type
func findItem(itemArr *gabs.Container, itemName string) []*GW2Item {
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
    //fmt.Println(charName)
    children, _ := chars.Children()
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
    //    for _, tuser := range users {
    //        fmt.Printf("username = %s GameId = %s Key = %s\n", tuser.ChatName, tuser.GameId, tuser.Key)
    //    }
    for _, user := range users {
        fmt.Println(user)
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

func getCacheAccountStruct(gw2 Gw2Api) *gabs.Container {
    var jsonParsed *gabs.Container

    value, ok := accountCache.age[gw2.Key]
    if (value+300 > time.Now().Unix()) && ok {
        fmt.Println("using cached values")
        jsonParsed = accountCache.accountCache[gw2.Key]
    } else {
        fmt.Println("getting new values")
        body := QueryAnetAuth(gw2, "account")
        jsonParsed, _ = gabs.ParseJSON(body)
        accountCache.accountCache[gw2.Key] = jsonParsed
        accountCache.age[gw2.Key] = time.Now().Unix()
    }

    return jsonParsed
}

func GetHomeWorld(gw2 Gw2Api) string {
    jsonParsed := getCacheAccountStruct(gw2)
    return jsonParsed.Search("world").String()
}

func GetWWWStats(gw2 Gw2Api, world string) GW2WvWvWStats {
    //: make(map[string]*gabs.Container), age: make(map[string]int64)}
    retVal := GW2WvWvWStats{}

    body := QueryAnet(gw2, "wvw/matches/stats", "world", world)
    jsonParsed, _ := gabs.ParseJSON(body)

    //fmt.Println(jsonParsed.String())
    //fmt.Println(jsonParsed.Path("deaths.red").Data().(float64))

    retVal.Deaths.Red = jsonParsed.Path("deaths.red").Data().(float64)
    retVal.Deaths.Blue = jsonParsed.Path("deaths.blue").Data().(float64)
    retVal.Deaths.Green = jsonParsed.Path("deaths.green").Data().(float64)

    retVal.Kills.Red = jsonParsed.Path("kills.red").Data().(float64)
    retVal.Kills.Blue = jsonParsed.Path("kills.blue").Data().(float64)
    retVal.Kills.Green = jsonParsed.Path("kills.green").Data().(float64)

    return retVal
}

func GetWorlds(gw2 Gw2Api, worlds string) Worlds {
    retVal := Worlds{WorldName: make(map[int64]string)}

    body := QueryAnet(gw2, "worlds", "ids", worlds)
    dec := json.NewDecoder(strings.NewReader(string(body[:])))
    dec.UseNumber()
    jsonParsed, _ := gabs.ParseJSONDecoder(dec)

    items, _ := jsonParsed.Children()
    for _, item := range items {
        key, _ := item.Path("id").Data().(json.Number).Int64()
        retVal.WorldName[key] = item.Path("name").Data().(string)
    }

    return retVal
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
    fmt.Println(findItem(items, "food"))
    return
}
*/
