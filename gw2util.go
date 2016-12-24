package gw2util

import (
	"fmt"
	"github.com/Jeffail/gabs"
	"strconv"
	"strings"
	"encoding/json"
	"log"
)

type Gw2Api struct {
	BaseUrl string
	Key     string
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
	body := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	return getCrafting(jsonParsed, name)
}

func FindItem(itemArr *gabs.Container, itemName string) ([]*GW2Item) {
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

func GetItems(gw2 Gw2Api, Ids []uint64) *gabs.Container {
	strIds := arrayToString(Ids, ",")
	body := QueryAnet(gw2, "items", "ids", strIds)
	jsonParsed, _ := gabs.ParseJSON(body)
	return jsonParsed
}

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

func main() {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	body := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	//fmt.Println(jsonParsed.StringIndent("","\t"))
	craftings := getCrafting(jsonParsed, "Notamik")
	log.Println(craftings[0])
	items := GetItems(gw2, GetItemIdsFromBags(jsonParsed, "nomitik"))
	fmt.Println(FindItem(items, "food"))
	return
}