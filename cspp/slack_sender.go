package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

var (
	watcher *fsnotify.Watcher
	mu      sync.Mutex
)

func watchDirectory(directoryPath string, done chan struct{}) {
	log.Debugln("(watchDirectory)", directoryPath)
	if err := watcher.Add(directoryPath); err != nil {
		log.Errorln("Error watching directory:", err)
		return
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Debugln("(watchingDirectory) watcher.Events channel closed. Exiting watchDirectory.")
				return
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				go handleNewFile(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				log.Debugln("(watchDirectory) watcher.Errors channel closed. Exiting watchDirectory.")
				return
			}
			log.Errorln("Error watching directory:", err)

		case <-done:
			log.Debugln("(watchDirectory) Received signal to exit. Exiting watchDirectory.")
			return
		}

	}
}

func handleNewFile(filePath string) {
	// if it's not an image, use the json file to see comment

	log.Debugln("(handleNewFile) filePath:", filePath)
	mu.Lock()
	defer mu.Unlock()

	if isImage(filePath) {
		return
	}

	if isJson(filePath) {
		log.Debugln("(handleNewfile) Found a json file", filePath)
		var j ImageInfo
		var err error
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Debugln("(handleNewFile) Can't read the file", filePath)
		}
		err = json.Unmarshal(content, &j)
		if err != nil {
			// move the file to the invalid folder
			log.Warnln("Unable to process " + filePath + " moving to invalid folder")
			log.Debugln("(handleNewFile) Json unmarhsalling error is ", err)
			return
		}
		handleImageFile(j)
		moveToDir(filePath, viper.GetString("processed_dir"))
	}
}

func sendKeyInDM(slackId string, key string) error {
	log.Debugln("(sendKeyInDM) slackId", slackId, "key", key)
	slack_token := viper.GetString("slack_token")
	base_url := viper.GetString("base_url")
	if base_url == "" {
		base_url = "<service address>"
	}
	api := slack.New(slack_token)
	msg := "You are on your way to :poop:posting!\n"
	msg += "Your <https://github.com/stahnma/mandatoryFun/tree/main/cspp|CSPP> API key is:  `" + key + "`" + ". Please keep it safe and do not share it with anyone. "
	msg += "In most cases, you can use your key with the entire CSPP service via something like the following command:\n"
	cmd := "curl -X POST \\\n "
	cmd += "-F \"image=@/path/to/file\" \\\n "
	cmd += "-F \"caption=String you want to with the picture\" \\\n "
	cmd += "-H \"X-API-KEY: $API_KEY\" \\\n "
	cmd += base_url + "/upload" + "\n"
	msg += "```" + cmd + "```"
	msg += "\n See " + base_url + "/usage for more details."

	_, _, err := api.PostMessage(slackId,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		return fmt.Errorf("error sending message to Slack user: %v", err)
	}
	return nil
}

func handleImageFile(j ImageInfo) {
	filePath := j.ImagePath
	if isImage(filepath.Base(filePath)) {
		err := uploadImageToSlack(j)
		if err != nil {
			log.Errorf("File %s not uploaded. Error: %v\n", filepath.Base(filePath), err)
			return
		}
		moveToDir(filePath, viper.GetString("processed_dir"))
	}
}

func getAuthor(apiKey string) string {
	log.Debugln("(getAuthor)" + apiKey)
	filename := viper.GetString("credentials_dir") + "/" + apiKey + ".json"
	ae, err := loadApiEntryFromFile(filename)
	if err != nil {
		log.Errorln("Unable to load api entry from file: ", apiKey, " error: ", err)
	}
	token := viper.GetString("slack_token")
	slackApi := slack.New(token)
	userInfo, err := slackApi.GetUserInfo(ae.SlackId)
	if err != nil {
		log.Errorln("Unable to retrieve user info from slack for user id: ", ae.SlackId, " error: ", err)
		return "Author Unknown (Error retrieving user info from slack)"
	}
	log.Debugln("(getAuthor) userInfo.profile.display_name_normalized", userInfo.Profile.DisplayNameNormalized)
	return userInfo.Profile.DisplayNameNormalized
}

func uploadImageToSlack(j ImageInfo) error {
	slack_token := viper.GetString("slack_token")
	slack_channel := viper.GetString("slack_channel")
	slackApi := slack.New(slack_token)
	filePath := j.ImagePath
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var comment string
	if j.Caption == "" {
		comment = "New image uploaded to Slack!"
	} else {
		comment = j.Caption
	}

	params := slack.FileUploadParameters{
		File:           filePath,
		Filename:       filepath.Base(filePath),
		Filetype:       "auto",
		Title:          getAuthor(j.ApiKey),
		Channels:       []string{slack_channel},
		InitialComment: comment,
	}

	_, err = slackApi.UploadFile(params)
	if err != nil {
		return err
	}
	return nil
}

func isImage(fileName string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif"}
	lowerCaseFileName := strings.ToLower(fileName)
	for _, ext := range extensions {
		if strings.HasSuffix(lowerCaseFileName, ext) {
			return true
		}
	}
	return false
}

func hasJsonExtension(fileName string) bool {
	extensions := []string{".json"}
	lowerCaseFileName := strings.ToLower(fileName)
	for _, ext := range extensions {
		if strings.HasSuffix(lowerCaseFileName, ext) {
			return true
		}
	}
	return false
}

// FIXME - handle sceanrio where a spare image is just in the upload dir
func handleSpareImage(filePath string) {
	log.Debugln("(handleSpareImage) filePath: ", filePath)
	mu.Lock()
	defer mu.Unlock()
	moveToDir(filePath, viper.GetString("discard_dir"))
}

func isJson(filename string) bool {
	log.Debugln("(isJson) filename:", filename)
	if hasJsonExtension(filename) {
		log.Debugln(filename + " Has Json extension")
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Debugln("(isJson) Unable to read file", filename)
			return false
		}
		// See if it's valid JSON
		var jsonData interface{}
		err = json.Unmarshal(content, &jsonData)
		if err != nil {
			log.Debugln("(isJson) Unable to parse json. Error is", err)
			moveToDir(filename, viper.GetString("discard_dir"))
			log.Infoln("Moved " + filename + " to discard directory. Invalid JSON file or schema.")
			return false
		}
		return true
	}
	log.Debugln("(isJson) " + filename + "Not a json file")
	return false
}
