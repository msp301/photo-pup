package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/msp301/photo-pup/internal/photoslibrary"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientId string
	Secret   string
}

type AuthCode struct {
	Code  string
	State string
}

func (authCode AuthCode) isValid() bool {
	if authCode.State != "state" {
		return false
	}

	if authCode.Code == "" {
		return false
	}

	return true
}

func getAuthCode(channel chan AuthCode) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		code := AuthCode{
			Code:  req.FormValue("code"),
			State: req.FormValue("state"),
		}

		channel <- code
	})
}

func listen(port int, channel chan AuthCode) {
	http.HandleFunc("/redirect", getAuthCode(channel))
	http.ListenAndServe(":3001", nil)
}

func main() {
	cfg := Config{}
	_, err := toml.DecodeFile("config.toml", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	config := &oauth2.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.Secret,
		RedirectURL:  "http://127.0.0.1:3001/redirect",
		Scopes:       []string{"https://www.googleapis.com/auth/photoslibrary.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.google.com/o/oauth2/auth",
			TokenURL:  "https://oauth2.googleapis.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	ctx := context.Background()
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	channel := make(chan AuthCode)
	go listen(3001, channel)

	fmt.Printf("Authorise your Google account by visiting: %v\n", authURL)

	authCode := <-channel

	if !authCode.isValid() {
		log.Fatal("Invalid authorization code")
	}

	token, err := config.Exchange(ctx, authCode.Code)
	if err != nil {
		log.Fatal(err)
	}

	client := config.Client(ctx, token)

	resp, err := client.Get("https://photoslibrary.googleapis.com/v1/albums")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	albums := photoslibrary.AlbumsList{}
	err = json.Unmarshal(body, &albums)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(albums)
}
