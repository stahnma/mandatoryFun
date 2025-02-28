package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/grandcat/zeroconf"
)

const elgatoAPIPath = "/elgato/lights"

var (
	version   string
	commit    string
	buildDate string
)

type LightState struct {
	Lights []struct {
		On int `json:"on"`
	} `json:"lights"`
}

// findElgatoLight discovers the first Elgato Light on the local network using mDNS.
// It returns the IP address of the found light or an error if no light is found.
func findElgatoLight() (string, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return "", fmt.Errorf("failed to initialize resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		if err := resolver.Browse(ctx, "_elg._tcp", "local.", entries); err != nil {
			log.Printf("failed to browse mDNS: %v", err)
		}
	}()

	for entry := range entries {
		if len(entry.AddrIPv4) > 0 {
			return entry.AddrIPv4[0].String(), nil
		}
	}

	return "", fmt.Errorf("no Elgato light found")
}

// getCurrentState retrieves the current on/off state of the Elgato Light.
// It returns 1 if the light is on, 0 if it is off, or an error if the state cannot be retrieved.
func getCurrentState(ip string) (int, error) {
	url := fmt.Sprintf("http://%s:9123%s", ip, elgatoAPIPath)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to get current state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var state LightState
	if err := json.Unmarshal(body, &state); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(state.Lights) > 0 {
		return state.Lights[0].On, nil
	}

	return 0, fmt.Errorf("no light state data available")
}

// toggleLight changes the Elgato Light's state from on to off or vice versa.
// It fetches the current state, determines the opposite state, and sends a request to update it.
func toggleLight(ip string) error {
	currentState, err := getCurrentState(ip)
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	newState := 1 - currentState

	url := fmt.Sprintf("http://%s:9123%s", ip, elgatoAPIPath)
	payload := LightState{
		Lights: []struct {
			On int `json:"on"`
		}{
			{On: newState},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}

	fmt.Printf("Light toggled to %s\n", map[int]string{0: "off", 1: "on"}[newState])
	return nil
}

// printHelp displays the usage instructions for the program.
func printHelp(programName string) {
	fmt.Printf(`Usage: %s [option]

This program finds the first Elgato Light on your network and toggles its state.

Options:
  --help         Show this help message.
  --version      Show program version.

Example usage:
  %s         Toggles the light on or off
  %s --help  Shows this message
`, programName, programName, programName)
}

// main is the entry point of the program. It detects Elgato Lights and toggles their state.
func main() {
	programName := filepath.Base(os.Args[0])

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help":
			printHelp(programName)
			return
		case "--version":
			fmt.Printf("Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, buildDate)
			return
		}
	}

	ip, err := findElgatoLight()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if err := toggleLight(ip); err != nil {
		log.Fatalf("Failed to toggle light: %v", err)
	}
}
