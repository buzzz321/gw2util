package gw2util

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs"
)

type GW2Crafting struct {
	Discipline string
	Rating     int64
	Active     bool
}

type GW2Item struct {
	ChatLink string `json:"chat_link"`
	Details  struct {
		ApplyCount  int     `json:"apply_count"`
		Description string  `json:"description"`
		DurationMs  float64 `json:"duration_ms"`
		Icon        string  `json:"icon"`
		Name        string  `json:"name"`
		Type        string  `json:"type"`
	} `json:"details"`
	Flags        []string `json:"flags"`
	GameTypes    []string `json:"game_types"`
	Icon         string   `json:"icon"`
	ID           int      `json:"id"`
	Level        int      `json:"level"`
	Name         string   `json:"name"`
	Rarity       string   `json:"rarity"`
	Restrictions []string `json:"restrictions"`
	Type         string   `json:"type"`
	VendorValue  int      `json:"vendor_value"`
}

func (b GW2Crafting) String() string {
	return fmt.Sprintf("\nDiscipline: %s\nRating %d \nActive %t\n", b.Discipline, b.Rating, b.Active)
}

func (b GW2Item) String() string {
	return fmt.Sprintf("\n Name: %s\n Type: %s \nDetail.Name: %s\n", b.Name, b.Type, b.Details.Name)
}
func GetKey(filename string) string {
	buff, err := ioutil.ReadFile(filename) // just pass the file name
	if err != nil {
		fmt.Print(err)
		return ""
	}

	return strings.Trim(string(buff), "\n ")
}

func arrayToString(a []uint64, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func flattenIdArray(objectOfIdArrays *gabs.Container) []uint64 {
	var (
		retVal []uint64
		tmpArr []uint64
	)

	arrayOfIdArrays, _ := objectOfIdArrays.Children()

	for _, IdArray := range arrayOfIdArrays {
		Ids, _ := IdArray.Children()
		length := len(IdArray.Data().([]interface{}))
		tmpArr = make([]uint64, length)
		for index, item := range Ids {
			tmpArr[index] = uint64(item.Data().(float64))
		}
		retVal = append(retVal, tmpArr...)
	}
	return retVal
}

func doRestQuery(Url string) []byte {
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

func QueryAnet(gw2 Gw2Api, command string, paramName string, paramValue string) []byte {
	Url := fmt.Sprintf("%s%s%s%s%s%s", gw2.BaseUrl, command, "?", paramName, "=", paramValue)
	fmt.Println("Url: ", Url)
	return doRestQuery(Url)
}

func QueryAnetAuth(gw2 Gw2Api, command string) []byte {
	Url := fmt.Sprintf("%s%s%s%s%s", gw2.BaseUrl, command, "?access_token=", gw2.Key, "&page=0")

	return doRestQuery(Url)
}
