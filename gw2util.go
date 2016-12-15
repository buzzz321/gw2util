package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"net/http"
	"log"
	"github.com/Jeffail/gabs"
	"strconv"
)

type Gw2Api struct {
	BaseUrl string
	Key     string
}
type Crafting struct {
	Discipline string
	Rating     int64
	Active     bool
}

func (b Crafting) String() string {
	return fmt.Sprintf("\nDiscipline: %s\nRating %d \nActive %t\n", b.Discipline, b.Rating, b.Active)
}

func getKey(filename string) string {
	buff, err := ioutil.ReadFile(filename) // just pass the file name
	if err != nil {
		fmt.Print(err)
		return ""
	}

	return strings.Trim(string(buff), "\n ")
}

func getCrafting(chars *gabs.Container, name string) ([]Crafting) {
	size, _ := chars.ArrayCount("name")
	var retVal = make([]Crafting, 0)

	for index := 0; index < size; index++ {
		if strings.Contains(chars.Index(index).Search("name").String(), name) {
			crafts := chars.Index(index).Search("crafting")
			discipline := crafts.Search("discipline")
			rating := crafts.Search("rating")
			active := crafts.Search("active")
			amountD, _ := crafts.ArrayCount("discipline")
			amountR, _ := crafts.ArrayCount("rating")
			amountA, _ := crafts.ArrayCount("active")

			if ( (amountD != amountR) && (amountD != amountA)) {
				return retVal
			}
			for n := 0; n < amountD; n++ {
				Discipline := discipline.Index(n).String()
				Rating, _ := strconv.ParseInt(rating.Index(n).String(), 10, 64)
				Active, _ := strconv.ParseBool(active.Index(n).String())
				retVal = append(retVal, Crafting{Discipline, Rating, Active})
			}
		}
	}

	return retVal
}

func flattenIdArray(objectOfIdArrays *gabs.Container) ([]uint64) {
	var (
		retVal  []uint64
		tmpArr []uint64
	)

	arrayOfIdArrays, _ := objectOfIdArrays.Children()

	for _, IdArray := range arrayOfIdArrays {
		Ids, _ := IdArray.Children()
		length := len(IdArray.Data().([] interface{}))
		tmpArr = make([]uint64, length)
		for index, item := range Ids {
			tmpArr[index] = uint64(item.Data().(float64))
		}
		retVal = append(retVal, tmpArr...)
	}
	return retVal
}

func arrayToString(a []uint64, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func getItems(gw2 Gw2Api, Ids [] uint64) (*gabs.Container) {
	strIds := arrayToString(Ids, ",")
	body := queryAnet(gw2, "items", "ids", strIds)
	jsonParsed, _ := gabs.ParseJSON(body)
	return jsonParsed
}

func getItemIdsFromBags(chars *gabs.Container, charName string) ([]uint64) {
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

func doRestQuery(Url string) ([] byte) {
	tr := &http.Transport{
		DisableCompression: true,
	}

	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return nil
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return nil
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return body
}
func queryAnet(gw2 Gw2Api, command string, paramName string, paramValue string) ([] byte) {
	Url := fmt.Sprintf("%s%s%s%s%s%s", gw2.BaseUrl, command, "?", paramName, "=", paramValue)
	fmt.Println("Url: ", Url)
	return doRestQuery(Url)
}
func queryAnetAuth(gw2 Gw2Api, command string) ([] byte) {
	Url := fmt.Sprintf("%s%s%s%s%s", gw2.BaseUrl, command, "?access_token=", gw2.Key, "&page=0")

	return doRestQuery(Url)
}
func main() {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", getKey("../../../gw2/test.key")}

	body := queryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	//fmt.Println(jsonParsed.StringIndent("","\t"))
	craftings := getCrafting(jsonParsed, "Notamik")
	log.Println(craftings[0])
	getItems(gw2, getItemIdsFromBags(jsonParsed, "nomitik"))
	return
}
