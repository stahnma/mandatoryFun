package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

type ApiEntry struct {
	ApiKey    string `json:"api_key"`
	IssueDate string `json:"issue_date"`
	LastUsed  string `json:"last_used"`
	Revoked   bool   `json:"revoked"`
	SlackId   string `json:"slack_id"`
}

type ApiKeyRequest struct {
	SlackId string `json:"slack_id"`
}

func (ae ApiEntry) save() {
	log.Debugln("(Save) Inside save")
	jsonData, err := json.Marshal(ae)
	if err != nil {
		log.Errorln("(Save) Error:", err)
		return
	}
	filename := ae.ApiKey + ".json"
	err = os.WriteFile(viper.GetString("credentials_dir")+"/"+filename, jsonData, 0o644)
	if err != nil {
		log.Warnln("(Save) Error writing to file:", err)
		return
	}
	log.Debugln("(Save) JSON data written to", filename)
}

func (ae ApiEntry) isRevoked() bool {
	return ae.Revoked
}

func (ae ApiEntry) updateLastUsed() {
	ae.LastUsed = time.Now().Format(time.RFC3339)
	ae.save()
}

func SearchAPIKeyInDirectory(searchString string) ([]string, error) {
	directoryPath := viper.GetString("credentials_dir")
	var matches []string
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			if match, err := searchAPIKeyInFile(path, searchString); err == nil && match {
				matches = append(matches, path)
			}
		}
		return nil
	})
	return matches, err
}

func searchAPIKeyInFile(filePath, searchString string) (bool, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}
	var ae ApiEntry
	err = json.Unmarshal(fileContent, &ae)
	if err != nil {
		return false, err
	}
	return ae.ApiKey == searchString, nil
}

// Fixme this is not quite right yet
func validateApiKey(key string) (bool, error) {
	log.Debugln("(validateApiKey)")
	// give a key, scan all files  and look for it
	matches, err := SearchAPIKeyInDirectory(key)
	if err != nil {
		log.Errorln("Error:", err)
		return false, nil
	}
	log.Debugln("(validateApiKey) Matches", matches)
	log.Debugln("(validateApiKey) Lenght of Matches", len(matches))
	if len(matches) < 1 {
		return false, nil
	}
	for _, match := range matches {
		log.Debugln("(validateApiKey) Match found in file:", match)
		isRev, err := isRevoked(match)
		if err != nil {
			log.Errorln("(validateApiKey) Error:", err)
			return false, nil
		}
		revErr := errors.New("(validateApiKey) Key has been revoked")
		if isRev {
			log.Debugln("(validateApiKey) Key has been revoked")
			return false, revErr
		}
	}
	return true, nil
}

func loadApiEntryFromFile(filePath string) (ApiEntry, error) {
	log.Debugln("(loadApiEntryFromFile)", filePath)
	var ae ApiEntry
	filecontent, err := os.ReadFile(filePath)
	if err != nil {
		log.Errorln("(loadApiEntryFromFile) Error reading file:"+filePath+" ", err)
		return ae, err
	}
	err = json.Unmarshal(filecontent, &ae)
	if err != nil {
		log.Errorln("(loadApiEntryFromFile) Error unmarshalling json from keyfile ", filePath, err)
		return ae, err
	}
	return ae, nil
}

// TODO get team id from a global var
func issueNewApiKey(slackId string) bool {
	log.Debugln("(issueNewApiKey) slackId", slackId)
	var keyBlob ApiEntry
	teamId := viper.GetString("slack_team_id")
	b := validateSlackId(slackId, teamId)
	log.Debugln("(issueNewApiKey) validateSlackId returned: ", b)
	// at this point we know the slack id is valid
	if b {
		keyBlob.IssueDate = time.Now().Format(time.RFC3339)
		keyBlob.ApiKey = generateApiKey()
		keyBlob.SlackId = slackId
		// TODO creatre revocation mechanism
		keyBlob.Revoked = false
		log.Debugln("(issueNewApiKey) keyBlob: ", keyBlob.ApiKey)
		keyBlob.save()
		err := sendKeyInDM(slackId, keyBlob.ApiKey)
		if err != nil {
			log.Errorln("(issueNewApiKey) Error sending key in DM:", err)
			return false
		}
		return true
	}
	return false
}

func revokeApiKey(key string) bool {
	log.Debugln("(revokeApiKey)", key)
	// find the key file
	// change json to revoked = true
	var ae ApiEntry
	keyfile := viper.GetString("credentials_dir") + "/" + key + ".json"
	filecontent, err := os.ReadFile(keyfile)
	if err != nil {
		log.Errorln("(revokeAPiKey) Error reading file:"+keyfile+" ", err)
		return false
	}
	// read file into json struct
	err = json.Unmarshal(filecontent, &ae)
	if err != nil {
		log.Errorln("(revokeApiKey) Error unmarshalling json from keyfile ", keyfile, err)
		return false
	}
	ae.Revoked = true
	ae.save()
	return true
}

func isRevoked(filePath string) (bool, error) {
	log.Debugln("(isRevoked)", filePath)
	ae, err := loadApiEntryFromFile(filePath)
	if err != nil {
		log.Errorln("(isRevoked) Error loading api entry from file", filePath, err)
		// This is fail-safe
		return true, err
	}
	log.Debugln("(isRevoked) ae.Revoked", ae.Revoked)
	if ae.Revoked == false {
		ae.updateLastUsed()
	}
	return ae.Revoked == true, nil
}

func generateApiKey() string {
	key := uuid.NewString()
	log.Debugln("(generateApiKey)", key)
	return key
}

func validateSlackId(userID, teamID string) bool {
	token := viper.GetString("slack_token")
	log.Debugln("(validateSlackId) userId: ", userID, " teamId: ", teamID, " token: ", token)
	api := slack.New(token)
	userInfo, err := api.GetUserInfo(userID)
	if err != nil {
		log.Errorln("Unable to retrieve user info from slack for user id: ", userID, " team id: ", teamID, " error: ", err)
		return false
	}
	log.Debugln("(validateSlackId) userInfo.TeamID", userInfo.TeamID)
	return userInfo.TeamID == teamID
}
