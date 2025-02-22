package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func parseAndSplitSlackHooks(msg string, jresp ZoomWebhook) {
	log.Debugln("(parseAndSplitSlackHooks) Parsing and splitting slack hooks.")
	log.Debugln("(parseAndSplitSlackHooks) The message is:", msg)
	log.Debugln("(parseAndSplitSlackHooks) The meeting ID is:", jresp.Payload.Object.ID)
	slackHooks := viper.GetString("slack_webhook_uri")

	splitStrings := strings.Split(slackHooks, ",")
	for i, s := range splitStrings {
		splitStrings[i] = strings.ReplaceAll(s, "'", "")
		splitStrings[i] = strings.ReplaceAll(s, "\"", "")
	}
	size := len(splitStrings)
	log.Debugln("(parseAndSplitSlackHooks) Found", size, "slack hooks.")
	msg = formatSlackMessage(msg, jresp)
	for _, entry := range splitStrings {
		postToSlack(msg, entry)
	}
}

func formatSlackMessage(msg string, jresp ZoomWebhook) string {
	// Read in suffix, and make that the hot link
	log.Debugln("XXXXXXX")
	log.Debugln("(formatSlackMessage) msg is ", msg)
	joinurl := callZoomAPI(jresp.Payload.Object.ID)
	log.Debugln("(formatSlackMessage) The join URL is:", joinurl)

	msg_suffix := viper.GetString("msg_suffix")
	log.Debugln("(formatSlackMessage) The suffix is:", msg_suffix)
	msg_suffix = "<" + joinurl + "|" + msg_suffix + ">"

	msg = msg + " " + msg_suffix
	log.Debugln("(formatSlackMessage) The message is:", msg)

	switch jresp.Event {
	case "meeting.participant_left":
		msg = jresp.Payload.Object.Participant.UserName + " has left " + msg_suffix
	case "meeting.participant_joined":
		msg = jresp.Payload.Object.Participant.UserName + " has joined " + msg_suffix
	default:
		msg = msg

	}
	return msg
}

func postToSlack(msg string, uri string) {
	log.Debugln("(postToSlack) The message is:", msg)
	log.Debugln("(postToSlack) The slack webhook uri is:", uri)

	// Constructing payload with Markdown and unfurl prevention
	payload := map[string]interface{}{
		"text":         msg,
		"mrkdwn":       true,
		"unfurl_links": false,
		"unfurl_media": false,
	}

	// Convert to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	// Make HTTP POST request
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Println("Error posting to Slack:", err)
		return
	}
	defer resp.Body.Close()

	log.Debugln("(postToSlack) Response Status:", resp.Status)
}
