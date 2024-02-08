package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/*
This next line is significant. It tells the compiler to include the static
file in the binary. There also cannot be space between // and go:embed.
*/

//go:embed static/usage.md
var md embed.FS

// ImageInfo represents the data to be stored in the JSON file
type ImageInfo struct {
	ImagePath string `json:"image_path"`
	Caption   string `json:"caption"`
	ApiKey    string `json:"api_key"`
}

func uploadAuthorization(c *gin.Context) error {
	log.Debugln("(uploadAuthorizion) Inside uploadAuthorizion")
	apiKey := c.GetHeader("X-API-Key")
	isApiKeyValied, err := validateApiKey(apiKey)
	if err != nil {
		log.Debug("(uploadHandler) validateApiKey threw error")
		if err.Error() != "Key has been revoked" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return err
		}
	}
	if isApiKeyValied {
		log.Debugln("(uploadHandler) API key is valid")
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Error(fmt.Errorf("Unauthorized request from IP: %s", c.ClientIP()))
		log.Warnln("Request has invalid API key")
		err = fmt.Errorf("Invalid API Key")
		return err
	}
	return nil
}

// FIXME Refactor this into shorter methods
func uploadHandler(c *gin.Context) {
	log.Debugln("(uploadHandler) Inside uploadHandler")
	apiKey := c.GetHeader("X-API-Key")
	if err := uploadAuthorization(c); err != nil {
		return
	}

	startTime := time.Now()

	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to parse form"})
		c.Error(fmt.Errorf("Error parsing form data: %v (Status: %d)", err, http.StatusBadRequest))
		return
	}

	image, imageHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to get image"})
		c.Error(fmt.Errorf("Error getting image file: %v (Status: %d)", err, http.StatusBadRequest))
		return
	}
	defer image.Close()

	caption := c.Request.FormValue("caption")
	imageName := fmt.Sprintf("%d_%s", getCurrentTimestamp(), imageHeader.Filename)
	uploadDir := viper.GetString("uploads_dir")
	setupDirectory(uploadDir)
	imagePath := filepath.Join(uploadDir, imageName)
	file, err := os.Create(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create file"})
		c.Error(fmt.Errorf("Error creating file: %v (Status: %d)", err, http.StatusInternalServerError))
		return
	}
	defer file.Close()

	// Save the image to disk
	_, err = io.Copy(file, image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save image"})
		c.Error(fmt.Errorf("Error saving image file: %v (Status: %d)", err, http.StatusInternalServerError))
		return
	}

	// TODO - remove the img extension from the json filename
	jsonPath := filepath.Join(uploadDir, fmt.Sprintf("%s.json", imageName))
	imageInfo := ImageInfo{ImagePath: imagePath, Caption: caption, ApiKey: apiKey}
	jsonData, err := json.MarshalIndent(imageInfo, "", "    ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create JSON"})
		c.Error(fmt.Errorf("Error creating JSON: %v (Status: %d)", err, http.StatusInternalServerError))
		return
	}
	err = os.WriteFile(jsonPath, jsonData, 0o644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save JSON"})
		c.Error(fmt.Errorf("Error saving JSON file: %v (Status: %d)", err, http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Upload successful", "filename": imageName})

	// Log the upload in the same format as [GIN] logging lines
	fmt.Fprintf(gin.DefaultWriter, "[GIN] %s | %3d | %13v | %15s | %-7s %s | Filename: %s\n",
		startTime.Format("2006/01/02 - 15:04:05"),
		http.StatusOK,
		time.Since(startTime),
		c.ClientIP(),
		c.Request.Method,
		c.Request.URL.Path,
		imageName,
	)
}

func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func receiver(done chan struct{}) {
	router := gin.New()
	router.POST("/upload", uploadHandler)
	router.POST("/api", postApiKeyHandler)
	router.DELETE("/api", deleteApiKeyHandler)
	router.GET("/usage", staticFileServer)
	// TODO implement the PUT method to replace the API key
	router.Run(":" + viper.GetString("PORT"))
}

func staticFileServer(c *gin.Context) {
	log.Debugln("(staticFileServer)")
	markdownContent, err := md.ReadFile("static/usage.md")
	if err != nil {
		log.Errorln("Error reading usage.md:", err)
		c.String(http.StatusInternalServerError, "Error reading static documentation")
		return
	}
	html := blackfriday.Run(markdownContent)
	c.Data(http.StatusOK, "text/html; charset=utf-8", html)
	return
}

func deleteApiKeyHandler(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if c.Request.Method == "DELETE" {
		log.Debugln("(deleteApiKeyHandler) DELETE request")
		good, err := validateApiKey(apiKey)
		if err != nil {
			log.Warnln("Error:", err)
			c.JSON(http.StatusNetworkAuthenticationRequired, gin.H{"status": "error", "message": err.Error()})
		}
		if good {
			revoked := revokeApiKey(apiKey)
			if revoked {
				c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Key revoked."})
			} else {
				log.Errorln("Unable to revoke key, but key file found. This is bad.")
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Unable to revoke key."})
			}
		} else {
			log.Warnln("Unablee to revoke key because key is not valid.", apiKey)
			c.JSON(http.StatusNetworkAuthenticationRequired, gin.H{"status": "error", "message": "Key not valid."})
		}
	}
}

func postApiKeyHandler(c *gin.Context) {
	if c.Request.Method == "POST" {
		var ae ApiKeyRequest
		if err := c.ShouldBindJSON(&ae); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Debugln("(postApiKeyHandler) Error processing JSON POST")
			return
		}
		log.Debugln("(postApiKeyHandler) ae.SlackId is:", ae.SlackId)
		slackId := ae.SlackId
		if issueNewApiKey(slackId) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "apikey": "See slack DM for your new API key."})
		} else {
			c.JSON(http.StatusNetworkAuthenticationRequired, gin.H{"status": "SlackID not found for team or error creating key."})
		}
	}
}
