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

func TestSetUserData(t *testing.T) {
	res := SaveUserData(UserDataSlice{UserData{"Chatter", "Gamer", "ASDFGHJKL"}})
	if res != "" {
		t.Errorf(res)
	}
	filename := "data.dat"
	data := ReadUserData(filename)
	if data[0].GameId != "Chatter" {
		t.Errorf("User not saved gameId = %s", data[0].GameId)
	}
	data = append(data, UserData{"Chatter2", "Gamer2", "2ASDFGHJKL"})
	res = SaveUserData(data)
	data = ReadUserData(filename)
	if data[0].GameId != "Chatter" {
		t.Errorf("User not saved gameId = %s", data[0].GameId)
	}
	if data[1].GameId != "Chatter2" {
		t.Errorf("User not saved expected Chatter2 got %s", data[1].GameId)
	}
	if res != "" {
		t.Errorf(res)
	}
}

func TestGetUserData(t *testing.T) {
	testData := UserDataSlice{UserData{GameId: "Chatter", ChatName: "Gamer", Key: "ASDFGHJKL"}, UserData{GameId: "Chatter2", ChatName: "Gamer2", Key: "ZXCVBN"}}

	test1 := GetUserData(testData, "Gamer2")
	if test1.GameId != "Chatter2" {
		t.Errorf("didnt find user expected Chatter2 got = %s", test1.ChatName)
	}
	test1 = GetUserData(testData, "Gamer")
	if test1.GameId != "Chatter" {
		t.Errorf("didnt find user expected Chatter got = %s", test1.ChatName)
	}
}

/*
func TestUpsertUserData(t *testing.T) {
	testData := UserDataSlice{UserData{GameId: "Chatter", ChatName: "Gamer", Key: "ASDFGHJKL"}, UserData{GameId: "Chatter2", ChatName: "Gamer2", Key: "ZXCVBN"}}

}
*/
