package gw2util

import (
	"fmt"
	"testing"

	"github.com/Jeffail/gabs"
)

func TestGetCharacterNames(t *testing.T) {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	body := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)
	chars := getCharacterNames(jsonParsed)

	fmt.Println(chars)
	/* some random test..
	   t.Errorf("...")

	*/
}

func TestSearchInBags(t *testing.T) {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	body := QueryAnetAuth(gw2, "characters")
	jsonParsed, _ := gabs.ParseJSON(body)

	items := GetItems(gw2, GetItemIdsFromBags(jsonParsed, "nomitik"))
	itemsMatch := findItem(items, "pistol")

	for _, item := range itemsMatch {
		fmt.Println(item.String())
	}
}

func TestGetHomeWorld(t *testing.T) {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	world := GetHomeWorld(gw2)
	fmt.Printf("home world = %s\n", world)

	if world != "2007" {
		t.Errorf("didnt get correct world expected 2007 got = %s", world)
	}
}

func TestGetWWWStats(t *testing.T) {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	fmt.Print(GetWWWStats(gw2, "2007"))
}

func TestGetWorlds(t *testing.T) {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	fmt.Printf("\nWorld test \n")
	fmt.Printf("%v\n", GetWorlds(gw2, "2007"))
}

func TestSetUserData(t *testing.T) {
	res := SaveUserData(UserDataSlice{UserData{"Chatter", "Gamer", "ASDFGHJKL"}})
	if res != "" {
		t.Errorf(res)
	}
	filename := "data.dat"
	data := ReadUserData(filename)
	if data[0].GameID != "Chatter" {
		t.Errorf("User not saved gameId = %s", data[0].GameID)
	}
	data = append(data, UserData{"Chatter2", "Gamer2", "2ASDFGHJKL"})
	res = SaveUserData(data)
	data = ReadUserData(filename)
	if data[0].GameID != "Chatter" {
		t.Errorf("User not saved gameId = %s", data[0].GameID)
	}
	if data[1].GameID != "Chatter2" {
		t.Errorf("User not saved expected Chatter2 got %s", data[1].GameID)
	}
	if res != "" {
		t.Errorf(res)
	}
}

func TestGetUserData(t *testing.T) {
	testData := UserDataSlice{UserData{GameID: "Chatter", ChatName: "Gamer", Key: "ASDFGHJKL"}, UserData{GameID: "Chatter2", ChatName: "Gamer2", Key: "ZXCVBN"}}

	test1 := GetUserData(testData, "Gamer2")
	if test1.GameID != "Chatter2" {
		t.Errorf("didnt find user expected Chatter2 got = %s", test1.ChatName)
	}
	test1 = GetUserData(testData, "Gamer")
	if test1.GameID != "Chatter" {
		t.Errorf("didnt find user expected Chatter got = %s", test1.ChatName)
	}
}

func TestGetWvWvWMatchParticipants(t *testing.T) {
	gw2 := Gw2Api{"https://api.guildwars2.com/v2/", GetKey("../../../gw2/test.key")}

	fmt.Printf("\nWorld test \n")
	fmt.Printf("%v\n", GetWvWvWColours(gw2, "2007"))
}

/*
func TestUpsertUserData(t *testing.T) {
    testData := UserDataSlice{UserData{GameId: "Chatter", ChatName: "Gamer", Key: "ASDFGHJKL"}, UserData{GameId: "Chatter2", ChatName: "Gamer2", Key: "ZXCVBN"}}

}
*/
