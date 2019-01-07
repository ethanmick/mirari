package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ethanmick/gathering"
	"github.com/fsnotify/fsnotify"
)

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

func upload(loc string) error {
	file, err := os.Open(loc)
	if err != nil {
		return err
	}
	defer file.Close()
	req, err := gathering.UploadFile("/upload/raw", "file", file)
	if err != nil {
		log.Printf("error creating file upload request: %v\n", err.Error())
		return err
	}
	var data interface{}
	_, err = gathering.Do(req, data)
	if err != nil {
		log.Printf("error uploading file: %v\n", err.Error())
		return nil
	}
	log.Printf("file upload success!")
	return nil
}

// ParseAll gets all data from a log
func ParseAll(f string) (gathering.UploadData, error) {
	data := gathering.UploadData{}
	str, err := readFile(f)
	if err != nil {
		return data, err
	}
	// Parse Everything
	col, err := gathering.ParseCollection(str)
	decks, err := gathering.ParseDecks(str)
	inv, err := gathering.ParsePlayerInventory(str)
	if err != nil {
		log.Printf("error parsing player inventory: %v\n", err.Error())
	}
	rank, err := gathering.ParseRankInfo(str)
	auth, err := gathering.ParseAuthRequest(str)
	matches := gathering.ParseMatches(str)
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
	if matches != nil {
		data.Matches = &matches
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
	req, err := gathering.Upload("/upload/json", body)
	if err != nil {
		log.Printf("error creating request: %v\n", err.Error())
		return
	}
	var data interface{}
	_, err = gathering.Do(req, data)
	if err != nil {
		log.Printf("error uploading data: %v\n", err.Error())
		return
	}
	log.Printf("upload success!")
}

func main() {
	log.Println("gathering client starting")
	var dirFlag = flag.String("dir", "", "The directory where the log file is located. This is useful when running on non-windows platforms where the log directory is not well known.")
	var filenameFlag = flag.String("file", fileName, "The name of the log file")
	var tokenFlag = flag.String("token", "", "Required: Your authentication token")
	var uploadFlag = flag.Bool("upload", true, "Upload raw file to server on start")
	flag.Parse()
	if *tokenFlag == "" {
		log.Fatalln("Error, need authentication token to upload data! Use `-token=TOKEN`")
	}
	gathering.Token = *tokenFlag
	user, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user: %v", err.Error())
	}
	log.Printf("user home directory: %v\n", user.HomeDir)
	var loc string
	if *dirFlag != "" {
		loc = *dirFlag
	} else {
		if gathering.LogDir == "" {
			log.Fatalf("Fatal: No log directory specified and the log location is unknown on this platform: '%v'. Please see `-help`\n", runtime.GOOS)
		}
		loc = filepath.Join(user.HomeDir, gathering.LogDir)
	}

	if *uploadFlag {
		upload(loc)
	}

	log.Printf("watching dir: '%v' with filename: '%v'\n", loc, *filenameFlag)
	watcher, err := fsnotify.NewWatcher()
	onChange(filepath.Join(loc, *filenameFlag), true)
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
