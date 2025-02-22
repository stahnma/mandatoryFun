package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	version   = "dev"  // Default to "dev" if not set
	commit    = "none" // Default to "none" if not set
	buildDate = "unknown"
)

type ZoomWebhook struct {
	Payload struct {
		PlainToken string `json:"plainToken"`
		AccountID  string `json:"account_id"`
		Object     struct {
			UUID        string `json:"uuid"`
			Participant struct {
				LeaveTime         time.Time `json:"leave_time"`
				JoinTime          time.Time `json:"join_time"`
				UserID            string    `json:"user_id"`
				UserName          string    `json:"user_name"`
				RegistrantID      string    `json:"registrant_id"`
				ParticipantUserID string    `json:"participant_user_id"`
				ID                string    `json:"id"`
				LeaveReason       string    `json:"leave_reason"`
				Email             string    `json:"email"`
				ParticipantUUID   string    `json:"participant_uuid"`
			} `json:"participant"`
			ID        string    `json:"id"`
			Type      int       `json:"type"`
			Topic     string    `json:"topic"`
			HostID    string    `json:"host_id"`
			Duration  int       `json:"duration"`
			StartTime time.Time `json:"start_time"`
			Timezone  string    `json:"timezone"`
		} `json:"object"`
	} `json:"payload"`
	EventTs int64  `json:"event_ts"`
	Event   string `json:"event"`
}

type ChallengeResponse struct {
	PlainToken     string `json:"plainToken"`
	EncryptedToken string `json:"encryptedToken"`
}

func zoomCrcValidation(jresp ZoomWebhook) (bool, ChallengeResponse) {
	log.Debugln("(zoomCrcValidation)")
	zoom_secret := viper.GetString("zoom_secret")
	var crc ChallengeResponse
	if jresp.Event == "endpoint.url_validation" {
		log.Debugln("(zoomCrcValidation) Performing CRC verification.")
		crc.PlainToken = jresp.Payload.PlainToken
		data := jresp.Payload.PlainToken
		// Create a new HMAC by defining the hash type and the key (as byte array)
		h := hmac.New(sha256.New, []byte(zoom_secret))
		h.Write([]byte(data))
		// Get result and encode as hexadecimal string
		crc.EncryptedToken = hex.EncodeToString(h.Sum(nil))
		log.Infoln("CRC Validation: ", crc)
		return true, crc
	} else {
		log.Debugln("(zoomCrcValidation) Not a CRC validation request.")
		return false, crc
	}

}

func filterMeeting(jresp ZoomWebhook) bool {
	// If the meeting is outside the topic scope, just ignore.
	name := viper.GetString("meeting_name")
	log.Debugln("(applyMeetingFilters) Topic " + jresp.Payload.Object.Topic)
	if name != jresp.Payload.Object.Topic && name != "" {
		log.Infoln("Received hook but dropping due to topic being filtered.")
		log.Debugln("(applyMeetingFilter) Hook had topic '" + jresp.Payload.Object.Topic + "'")
		log.Debugln("(applyMeetingFtiler)Filter only allows for " + name)
		return true
	}
	return false
}

func setMessageSuffix(jresp ZoomWebhook) string {
	msg_suffix := viper.GetString("msg_suffix")
	msg := ""
	switch jresp.Event {
	case "meeting.participant_left":
		msg = jresp.Payload.Object.Participant.UserName + " has left " + msg_suffix
	case "meeting.participant_joined":
		msg = jresp.Payload.Object.Participant.UserName + " has joined " + msg_suffix
	default:
		return msg
	}
	return msg
}

// savePostRequestToFile writes the payload of a POST request to a file in the "hooks" directory.
//
// It operates silently, meaning it does not modify the HTTP response or interfere with request processing.
// If the request is not a POST, has no body, or if an error occurs while reading or writing the body,
// the function simply returns without taking any action.
//
// The request payload is saved in a file named "hook-<timestamp>.json", where <timestamp> follows the
// "YYYYMMDD-HHMMSS.mmm" format to ensure unique filenames. The directory "hooks" is created if it does not exist.
//
// Parameters:
//   - c: A *gin.Context representing the HTTP request and response context.
//
// Usage:
//
//	This function can be used as middleware in a Gin application to capture incoming POST request payloads.
//
// Example:
//
//	r := gin.Default()
//	r.Use(savePostRequestToFile)  // Logs POST request bodies to the "hooks" directory
//	r.POST("/hook", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"message": "Hook received"})
//	})
//	r.Run(":8080")
func savePostRequestToFile(c *gin.Context) {
	if c.Request.Method != "POST" {
		return // Only process POST requests
	}

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		return // If there's an error reading or body is empty, do nothing
	}

	// Ensure the hooks directory exists
	hooksDir := "hooks"
	_ = os.MkdirAll(hooksDir, 0755)

	// Generate a timestamp-based filename
	timestamp := time.Now().Format("20060102-150405.000")
	filename := filepath.Join(hooksDir, "hook-"+timestamp+".json")

	// Write the body to a file (ignoring errors to avoid interfering with request processing)
	_ = os.WriteFile(filename, body, 0644)
}

func processWebHook(c *gin.Context) {

	log.Debugln("(processWebHook) Processing incoming webhook. at " + time.Now().String())
	if gin.IsDebugging() {
		// savePostRequestToFile(c)
	}
	var jresp ZoomWebhook
	if err := c.BindJSON(&jresp); err != nil {
		log.Errorln("Error processing incoming webhook JSON", err)
	}

	// Handle Zoom Webhook CRC validation
	if jresp.Event == "endpoint.url_validation" {
		crcvalid, crc := zoomCrcValidation(jresp)
		if crcvalid {
			log.Debugln("(processWebHook) CRC validation successful. Returning CRC response.")
			c.JSON(http.StatusOK, crc)
			return
		} else {
			log.Errorln("(processWebHook) CRC validation failed. Returning 400.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "CRC validation failed"})
			return
		}
	}
	if filterMeeting(jresp) {
		return
	}

	// If the event type is not a participant join or leave, ignore it.
	if jresp.Event != "meeting.participant_joined" && jresp.Event != "meeting.participant_left" {
		log.Debugln("(processWebHook) Ignoring event type: " + jresp.Event)
		return
	}

	msg := setMessageSuffix(jresp)
	log.Debugln("About to dispatch Message: " + msg)
	dispatchMessage(msg, jresp)
}

func getJoinURL(meetingId string) string {
	/*
			To enable this feature, you must set the following environment variables:
			ZOOM_API_ENABLE=1
		  ZOOM_API_CLIENT_ID
			ZOOM_API_CLIENT_SECRET
			ZOOM_API_ACCOUNT_ID
	*/
	joinurl := ""
	if os.Getenv("ZOOM_API_ENABLE") == "1" {
		// Get secret from the zoom API so we can get the meeting details
		// Check to see that ZOOM_API_CLIENT_ID, ZOOM_API_CLIENT_SECRET, and ZOOM_API_ACCOUNT_ID are set
		if os.Getenv("ZOOM_API_CLIENT_ID") != "" && os.Getenv("ZOOM_API_CLIENT_SECRET") != "" && os.Getenv("ZOOM_API_ACCOUNT_ID") != "" {
			log.Debugln(meetingId)
			joinurl := callZoomAPI(meetingId)
			log.Debugln("Join URL: " + joinurl)
		} else {
			// This should be unreachable code, but it's there for debugging and defense.
			log.Errorln("ZOOM_API environment credentials are not set. Skipping.")
		}
	}
	return joinurl
}

func dispatchMessage(msg string, jresp ZoomWebhook) {

	slack_enable := viper.GetString("slack_enable")
	irc_enable := viper.GetString("irc_enable")
	log.Debugln("(dispatchMessage) Slack enabled: " + slack_enable)
	log.Debugln("(dispatchMessage) IRC enabled: " + irc_enable)
	sent := 0

	if strings.ToLower(slack_enable) == "true" {
		log.Debugln("(dispatchMessage) Sending a slack message")
		parseAndSplitSlackHooks(msg, jresp)
		sent = 1
	}
	if strings.ToLower(irc_enable) == "true" {
		log.Debugln("(dispatchMessage) Sending an IRC message")
		sendIRC(msg)
		sent = 1
	}
	if sent == 0 {
		log.Fatal("You have no dispatchers configured (irc or slack). Quitting.")
	}
}

func inititalize() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	viper.SetConfigType("env")

	viper.SetDefault("port", "8888")
	viper.SetDefault("slack_enable", "true")
	viper.SetDefault("irc_enable", "false")
	viper.SetDefault("msg_suffix", "the zoom meeting.")
	viper.SetDefault("zoom_api_enable", "false")

	viper.BindEnv("port", "ZOOMWH_PORT")
	viper.BindEnv("slack_enable", "ZOOMWH_SLACK_ENABLE")

	bugout := false
	if value := os.Getenv("ZOOM_SECRET"); value == "" {
		bugout = true
		log.Errorln("You must set ZOOM_SECRET environment variable.")
	} else {
		viper.BindEnv("zoom_secret", "ZOOM_SECRET")
	}

	// Zoom API Specifics
	zoomApiEnabled := viper.GetBool("zoom_api_enable")
	if zoomApiEnabled == false {
		log.Infoln("Zoom Web API is disabled. Disabling active meeting links and quieries")
		viper.Set("zoom_api_enable", "false")
	} else {
		viper.MustBindEnv("zoom_api_client_id", "ZOOM_API_CLIENT_ID")
		zoom_api_client_id := viper.GetString("zoom_api_client_id")
		if zoom_api_client_id == "" {
			log.Errorln("You must set ZOOM_API_CLIENT_ID environment variable if ZOOM_API_ENABLE=true.")
			bugout = true
		}
		viper.MustBindEnv("zoom_api_client_secret", "ZOOM_API_CLIENT_SECRET")
		zoom_api_client_secret := viper.GetString("zoom_api_client_secret")
		if zoom_api_client_secret == "" {
			log.Errorln("You must set ZOOM_API_CLIENT_SECRET environment variable if ZOOM_API_ENABLE=true.")
			bugout = true
		}
		viper.MustBindEnv("zoom_api_account_id", "ZOOM_API_ACCOUNT_ID")
		zoom_api_account_id := viper.GetString("zoom_api_account_id")
		if zoom_api_account_id == "" {
			log.Errorln("You must set ZOOM_API_ACCOUNT_ID environment variable if ZOOM_API_ENABLE=true.")
			bugout = true
		}
	}

	// Slack Specifics
	viper.GetString("slack_enable")
	if value := os.Getenv("ZOOMWH_SLACK_ENABLE"); value == "false" {
		log.Infoln("Slack is notification disabled.")
		viper.Set("slack_enable", "false")
	} else {
		viper.MustBindEnv("slack_webhook_uri", "ZOOMWH_SLACK_WH_URI")
		slack_webhook_uri := viper.GetString("slack_webhook_uri")
		if slack_webhook_uri == "" {
			log.Errorln("You must set ZOOMWH_SLACK_WH_URI environment variable unless ZOOMWH_SLACK_ENABLE=false.")
			bugout = true
		}
	}

	// Filter Specifics
	if value := os.Getenv("ZOOMWH_MEETING_NAME"); value == "" {
		viper.BindEnv("meeting_filter", "ZOOMWH_MEETING_NAME")
	}

	// IRC Specifics
	value, ok := os.LookupEnv("ZOOMWH_IRC_ENABLE")
	if value == "false" || !ok {
		log.Infoln("IRC notifications are disabled.")
		viper.Set("irc_enable", "false")
	} else {
		log.Infoln("IRC notifications are enabled.")
		viper.Set("irc_enable", "true")
		// Four IRC variables are required if IRC is enabled
		if value := os.Getenv("ZOOMWH_IRC_SERVER"); value == "" {
			log.Errorln("You must set ZOOMWH_IRC_SERVER environment variable if ZOOMWH_IRC_ENABLE is true.")
			bugout = true
		} else {
			viper.MustBindEnv("irc_server", "ZOOMWH_IRC_SERVER")
		}
		if value := os.Getenv("ZOOMWH_IRC_CHANNEL"); value == "" {
			log.Errorln("You must set ZOOMWH_IRC_CHANNEL environment variable if ZOOMWH_IRC_ENABLE is true.")
			bugout = true
		} else {
			viper.MustBindEnv("irc_channel", "ZOOMWH_IRC_CHANNEL")
		}
		if value := os.Getenv("ZOOMWH_IRC_NICK"); value == "" {
			log.Errorln("You must set ZOOMWH_IRC_NICK environment variable if ZOOMWH_IRC_ENABLE is true.")
			bugout = true
		} else {
			viper.MustBindEnv("irc_nick", "ZOOMWH_IRC_NICK")
		}
		if value := os.Getenv("ZOOMWH_IRC_PASS"); value == "" {
			log.Errorln("You must set ZOOMWH_IRC_PASS environment variable if ZOOMWH_IRC_ENABLE is true.")
			bugout = true
		} else {
			viper.MustBindEnv("irc_pass", "ZOOMWH_IRC_PASS")
		}
	}

	// viper dump
	fmt.Println(viper.AllSettings())

	viper.MustBindEnv("zoom_secret", "ZOOM_SECRET")
	if os.Getenv("ZOOMWH_MSG_SUFFIX") != "" {
		viper.BindEnv("msg_suffix", "ZOOMWH_MSG_SUFFIX")
	}

	if bugout == true {
		os.Exit(1)
	}
}

func main() {

	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, buildDate)
		return
	}

	inititalize()

	router := gin.Default()
	router.POST("/", processWebHook)
	port := viper.GetString("port")
	serverstring := "localhost:" + port
	log.Infoln("Listening on " + serverstring)
	router.Run(serverstring)
}
