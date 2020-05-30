package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

func clean(str string) string {
	return strings.ReplaceAll(str, "/", "-")
}

func download(client *http.Client, filePath string, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func saveMediaItems(client *http.Client, media photoslibrary.MediaItemsList, outDir string) {
	for _, item := range media.MediaItems {
		downloadURL := item.BaseURL + "=d"
		outputFile := filepath.Join(outDir, item.Filename)

		if _, err := os.Stat(outputFile); err == nil {
			continue
		}

		err := download(client, outputFile, downloadURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func fetchAlbum(client *http.Client, albumID string, dir string) {
	nextPage := true
	query := url.Values{"albumId": {albumID}}

	for ok := true; ok; ok = nextPage {
		resp, err := client.PostForm("https://photoslibrary.googleapis.com/v1/mediaItems:search", query)
		if err != nil {
			log.Fatal(err)
		}

		media := photoslibrary.MediaItemsList{}
		err = json.NewDecoder(resp.Body).Decode(&media)
		if err != nil {
			log.Fatal(err)
		}

		saveMediaItems(client, media, dir)

		if media.NextPageToken != "" {
			query = url.Values{"albumId": {albumID}, "pageToken": {media.NextPageToken}}
		} else {
			nextPage = false
		}
	}
}

func fetchMediaItems(client *http.Client, albums []photoslibrary.Album) {
	for _, album := range albums {
		albumDir := clean(album.Title)

		if albumDir != "" {
			if _, err := os.Stat(albumDir); os.IsNotExist(err) {
				err := os.Mkdir(albumDir, os.ModeDir|0755)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		fetchAlbum(client, album.ID, albumDir)
	}
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

	albums := photoslibrary.AlbumsList{}
	err = json.NewDecoder(resp.Body).Decode(&albums)
	if err != nil {
		log.Fatal(err)
	}

	fetchMediaItems(client, albums.Albums)

	resp, err = client.Get("https://photoslibrary.googleapis.com/v1/sharedAlbums")
	if err != nil {
		log.Fatal(err)
	}

	sharedAlbums := photoslibrary.SharedAlbumsList{}
	err = json.NewDecoder(resp.Body).Decode(&sharedAlbums)
	if err != nil {
		log.Fatal(err)
	}

	fetchMediaItems(client, sharedAlbums.SharedAlbums)
}
