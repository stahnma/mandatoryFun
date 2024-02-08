package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func setupDirectory(directoryName string) {
	// if directoryName doesn't start with / file abosolute path'

	/* _, err := filepath.Abs(directoryName)
	if err != nil {
		// TODO do something with a DATA_DIR and build it from there.
	}
	*/
	var err error
	err = os.MkdirAll(directoryName, 0o755)
	if err != nil {
		log.Fatal("Error creating "+directoryName+" directory: ", err)
	}
}

func moveToDir(filePath, directoryName string) {
	destPath := filepath.Join(directoryName, filepath.Base(filePath))
	var err error
	err = os.Rename(filePath, destPath)
	if err != nil {
		log.Errorln("Error moving file " + filePath + " to " + directoryName + ": " + err.Error())
	} else {
		t := time.Now()
		fmt.Printf("[SND] %s %s sent to Slack and moved to \"processed\" directory.\n", t.Format("2006/01/02 - 15:04:05"), filepath.Base(filePath))
	}
}

func prettyPrintJSON(data interface{}) error {
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	log.Println(string(prettyJSON))
	return nil
}
