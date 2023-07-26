package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

func processWebHook(c *gin.Context) {

	if gin.IsDebugging() {
		// log incoming request if in DEBUG mode
	}

	var jresp ZoomWebhook
	if err := c.BindJSON(&jresp); err != nil {
		// DO SOMETHING WITH THE ERROR
		fmt.Println("Error processing JSON.", err)
	}

	// Handle Zoom Webhook CRC validation
	if jresp.Payload.PlainToken != "" {
		var crc ChallengeResponse
		crc.PlainToken = jresp.Payload.PlainToken
		//todo load this from env var
		zoom_secret := os.Getenv("ZOOM_SECRET")
		data := jresp.Payload.PlainToken
		// Create a new HMAC by defining the hash type and the key (as byte array)
		h := hmac.New(sha256.New, []byte(zoom_secret))
		h.Write([]byte(data))
		// Get result and encode as hexadecimal string
		crc.EncryptedToken = hex.EncodeToString(h.Sum(nil))
		fmt.Println(crc)
		c.JSON(http.StatusOK, crc)
		return
	}

	var msg string
	switch jresp.Event {
	case "meeting.participant_left":
		msg = jresp.Payload.Object.Participant.UserName + " has left the zoom meeting."
	case "meeting.participant_joined":
		msg = jresp.Payload.Object.Participant.UserName + " has joined the zoom meeting."
	default:
		return
	}
	postToSlackWebHook(msg)
	sendIRC(msg)
}

func postToSlackWebHook(msg string) {

	slack_uri := os.Getenv("ZOOM_WH_SLACK_WH_URI")
	data := url.Values{
		"payload": {"{\"text\": \"" + msg + "\"}"},
	}
	resp, err := http.PostForm(slack_uri, data)
	fmt.Println(resp.Status)
	if err != nil {
		log.Fatal(err)
	}
}

func validateEnvVars(key string) {
	_, ok := os.LookupEnv(key)
	if !ok {
		log.Fatal("You must set " + key + " environment variable.")
	}
}

func main() {

	validateEnvVars("ZOOM_SECRET")
	validateEnvVars("ZOOM_WH_SLACK_WH_URI")

	router := gin.Default()
	//TODO make this a configuration mount point
	router.POST("/", processWebHook)
	router.Run("localhost:8888")
}
