package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	webhookPort = "8080" // Change this port to the desired port number
	dataFile    = "webhook_data.json"
)

type WebhookData struct {
	// Define the structure of your JSON data here
	// For example:
	// Field1 string `json:"field1"`
	// Field2 int    `json:"field2"`
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var data WebhookData
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Failed to parse JSON data", http.StatusBadRequest)
		return
	}

	err = saveDataToFile(body)
	if err != nil {
		http.Error(w, "Failed to save data to file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Webhook data received and saved successfully!\n")
}

func saveDataToFile(data []byte) error {
	file, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	http.HandleFunc("/", handleWebhook)

	fmt.Printf("Webhook server listening on port %s...\n", webhookPort)
	err := http.ListenAndServe(":"+webhookPort, nil)
	if err != nil {
		fmt.Printf("Error starting webhook server: %s\n", err)
	}
}

