package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
)

type Response struct {
	Success bool   `json:"success"`
	User    []User `json:"data"`
}

type User struct {
	Id string `json:"_id"`
}

type InviteRequest struct {
	Uuids []string `json:"uuids"`
}

var apiUser string
var apiKey string

func main() {
	flag.StringVar(&apiUser, "api-user", "", "Habitica API user")
	flag.StringVar(&apiKey, "api-key", "", "Habitica API key")
	flag.Parse()

	if apiUser == "" || apiKey == "" {
		log.Fatal("Please provide Habitica API user and key. (Use -api-user and -api-key flags)")
	}

	fmt.Println("Welcome to PartyUp! The script will now start fetching users and inviting them to party.")
	go executeCronJob()
	time.Sleep(168 * time.Hour)
}

func executeCronJob() {
	gocron.Every(2).Minute().Do(fetchUsersAndInvite)
	<-gocron.Start()
}

func fetchUsersAndInvite() {
	fmt.Println("Fetching users and inviting them to party...")
	url := "https://habitica.com/api/v3/looking-for-party"

	habiticaClient := http.Client{
		Timeout: time.Second * 120,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-client", fmt.Sprintf("%s-PartyUp", apiUser))
	req.Header.Set("x-api-user", apiUser)
	req.Header.Set("x-api-key", apiKey)

	res, getErr := habiticaClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	var response Response
	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	if !response.Success {
		log.Fatal("Request failed")
	}

	usersUuid := make([]string, len(response.User))
	for _, user := range response.User {
		if user.Id != "" {
			usersUuid = append(usersUuid, user.Id)
		}
	}

	if len(usersUuid) != 0 {
		inviteUsers(habiticaClient, usersUuid)
	}
}

func inviteUsers(client http.Client, uuids []string) {
	inviteUrl := "https://habitica.com/api/v3/groups/party/invite"

	nonEmptyUuids := removeEmptyStrings(uuids)

	inviteRequest := InviteRequest{
		Uuids: nonEmptyUuids,
	}

	inviteBody, jsonErr := json.Marshal(inviteRequest)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	req, err := http.NewRequest(http.MethodPost, inviteUrl, bytes.NewBuffer(inviteBody))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-client", fmt.Sprintf("%s-PartyUp", apiUser))
	req.Header.Set("x-api-user", apiUser)
	req.Header.Set("x-api-key", apiKey)

	res, postErr := client.Do(req)
	if postErr != nil {
		log.Fatal(postErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	_, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	log.Println("All users invited!")
}

func removeEmptyStrings(input []string) []string {
	var result []string
	for _, str := range input {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}