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
	Key 	string
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
	var retVal = make([]Crafting, 10)

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
				retVal[n].Discipline = discipline.Index(n).String()
				retVal[n].Rating, _ = strconv.ParseInt(rating.Index(n).String(), 10, 64)
				retVal[n].Active, _ = strconv.ParseBool(active.Index(n).String())
			}
		}
	}

	return retVal
}

func queryData(gw2 Gw2Api, command string) ([] byte) {
	Url := fmt.Sprintf("%s%s%s%s%s", gw2.BaseUrl, command, "?access_token=", gw2.Key, "&page=0")
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
func main() {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", getKey("../../../gw2/test.key")}

	body := queryData(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	craftings := getCrafting(jsonParsed, "Notamik")
	log.Println(craftings[0])

	return
}
