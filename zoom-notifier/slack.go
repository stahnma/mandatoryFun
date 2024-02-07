package main

import (
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func parseAndSplitSlackHooks(msg string) {
	log.Debugln("(parseAndSplitSlackHooks) Parsing and splitting slack hooks.")
	slackHooks := viper.GetString("slack_webhook_uri")

	splitStrings := strings.Split(slackHooks, ",")
	for i, s := range splitStrings {
		splitStrings[i] = strings.ReplaceAll(s, "'", "")
		splitStrings[i] = strings.ReplaceAll(s, "\"", "")
	}
	size := len(splitStrings)
	log.Debugln("(parseAndSplitSlackHooks) Found", size, "slack hooks.")
	for _, entry := range splitStrings {
		postToSlack(msg, entry)
	}
}

func postToSlack(msg string, uri string) {
	log.Debugln("(postToSlack) Posting to each slack hook.", uri, msg)
	data := url.Values{
		"payload": {"{\"text\": \"" + msg + "\"}"},
	}
	resp, err := http.PostForm(uri, data)
	if err != nil {
		log.Errorln("Error posting to slack:", err)
	}
	log.Debugln(resp.Status)
}
