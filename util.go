package gw2util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

var gw2ApiError = []byte("\"text\": \"ErrInternal\"")

// GW2WvWvWStats is struct contains all kill and death on a border.
// ID is border id
type GW2WvWvWStats struct {
	ID   string
	Name string

	Deaths struct {
		Red   float64
		Blue  float64
		Green float64
	}

	Kills struct {
		Red   float64
		Blue  float64
		Green float64
	}
}

// GW2Crafting contains crafting info for a character
// Discipline is name of craft
// Rating     skill level of craft
// Active     tru if is currently active
type GW2Crafting struct {
	Discipline string
	Rating     int64
	Active     bool
}

// GW2Item used to unjasonify gw2item jason struct from server
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
	weblink := "https://wiki.guildwars2.com/wiki/" + strings.Replace(b.Name, " ", "_", -1)
	return fmt.Sprintf("\n Name: %s\n Type: %s \nChat Link: %s\n\nUrl: %s\n", b.Name, b.Type, b.ChatLink, weblink)
}

func (b GW2WvWvWStats) String() string {
	return fmt.Sprintf("\nWorld Name %s\n Deaths\n Red: %6.f\n Green: %6.f \n Blue: %6.f\n Kills\n Red: %6.f\n Green: %6.f \n Blue: %6.f\n",
		b.Name, b.Deaths.Red, b.Deaths.Green, b.Deaths.Blue, b.Kills.Red, b.Kills.Green, b.Kills.Blue)
}

// GetKey reads all file data and restusn it as string
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

func flattenIDArray(objectOfIDArrays *gabs.Container) []uint64 {
	var (
		retVal []uint64
		tmpArr []uint64
	)

	arrayOfIDArrays := objectOfIDArrays.Children()

	for _, IDArray := range arrayOfIDArrays {
		Ids := IDArray.Children()
		length := len(IDArray.Data().([]interface{}))
		tmpArr = make([]uint64, length)
		for index, item := range Ids {
			tmpArr[index] = uint64(item.Data().(float64))
		}
		retVal = append(retVal, tmpArr...)
	}
	return retVal
}

func doRestQuery(URL string) ([]byte, error) {
	tr := &http.Transport{
		DisableCompression: true,
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return nil, err
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return nil, err
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if len(body) < 10 {
		err = errors.New("ReadAll: rest request returned no data")
	}
	if err != nil {
		log.Fatal("Readall: ", err)
	}
	return body, err
}

// QueryAnet send web request to anet that dont require access token
func QueryAnet(gw2 Gw2Api, command string, paramName string, paramValue string) ([]byte, error) {
	URL := fmt.Sprintf("%s%s%s%s%s%s", gw2.BaseURL, command, "?", paramName, "=", paramValue)
	//fmt.Println("Url: ", URL)
	retry := false
	var queryRes []byte
	var err error
	tries := 0
	for !retry || tries > 10 {
		queryRes, err = doRestQuery(URL)
		if !bytes.Contains(queryRes, gw2ApiError) {
			retry = true
		}
		tries++
	}
	return queryRes, err
}

// QueryAnetAuth send web request to anet that require access token
func QueryAnetAuth(gw2 Gw2Api, command string) ([]byte, error) {
	URL := fmt.Sprintf("%s%s%s%s%s", gw2.BaseURL, command, "?access_token=", gw2.Key, "&page=0")

	retry := false
	var queryRes []byte
	var err error
	tries := 0
	for !retry || tries > 10 {
		queryRes, err = doRestQuery(URL)
		if !bytes.Contains(queryRes, gw2ApiError) {
			retry = true
		}
		tries++
	}
	return queryRes, err
}
