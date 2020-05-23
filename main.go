package main

import (
	"context"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientId string
	Secret   string
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
	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Printf("Authorise your Google account by visiting: %v\n", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	client := config.Client(ctx, token)

	albums, err := client.Get("https://photoslibrary.googleapis.com/v1/albums")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(albums)
}
