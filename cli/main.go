package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ethanmick/mirari"
	"github.com/fsnotify/fsnotify"
)

const logLocation = "\\AppData\\LocalLow\\Wizards Of The Coast\\MTGA"
const fileName = "output_log.txt"

var lastUpload = time.Now()

func readFile(f string) (string, error) {
	file, err := os.Open(f)
	if err != nil {
		return "", err
	}
	raw, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(raw[:]), nil
}

// ParseAll gets all data from a log
func ParseAll(f string) (mirari.UploadData, error) {
	data := mirari.UploadData{}
	str, err := readFile(f)
	if err != nil {
		return data, err
	}
	// Parse Everything
	col, err := mirari.ParseCollection(str)
	decks, err := mirari.ParseDecks(str)
	inv, err := mirari.ParsePlayerInventory(str)
	rank, err := mirari.ParseRankInfo(str)
	auth, err := mirari.ParseAuthRequest(str)
	// Store it all
	if col != nil {
		data.Collection = &col
	}
	if decks != nil {
		data.Decks = &decks
	}
	if inv != nil {
		data.Inventory = inv
	}
	if rank != nil {
		data.Rank = rank
	}
	if auth != nil {
		data.Auth = auth
	}
	return data, err
}

// onChange parse out all info and upload to the server
// Only uploads once a minute, unless forced
func onChange(f string, force bool) {
	if !force && lastUpload.Sub(time.Now()).Seconds() < 60 {
		log.Println("file changed but uploaded too recently, skipping")
		return
	}
	body, err := ParseAll(f)
	if err != nil {
		log.Printf("error parsing log file: %v\n", err.Error())
	}
	log.Println("uploading body (even if error)")
	req, err := mirari.Upload("/upload/raw", body)
	if err != nil {
		log.Printf("error creating request: %v\n", err.Error())
		return
	}
	var data interface{}
	_, err = mirari.Do(req, data)
	if err != nil {
		log.Printf("error uploading data: %v\n", err.Error())
		return
	}
	log.Printf("upload success!")
}

func main() {
	log.Println("mirari client starting")
	user, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user: %v", err.Error())
	}
	log.Printf("user home directory: %v\n", user.HomeDir)
	loc := filepath.Join(user.HomeDir, logLocation)
	log.Printf("watching dir: %v\n", loc)
	watcher, err := fsnotify.NewWatcher()
	onChange(filepath.Join(loc, fileName), true)
	if err != nil {
		log.Fatalf("error creating watcher: %v\n", err.Error())
	}
	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					if event.Name == filepath.Join(loc, fileName) {
						onChange(event.Name, false)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watcher error:", err)
			}
		}
	}()
	if err := watcher.Add(loc); err != nil {
		log.Fatalf("failed to watch directory: %v", err.Error())
	}
	<-done
}
