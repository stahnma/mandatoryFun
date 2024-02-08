package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// Set up Viper to use command line flags
	pflag.String("config", "", "config file (default is $HOME/.your_app.yaml)")
	pflag.Bool("ready-systemd", false, "flag to indicate systemd readiness")

	// Bind the command line flags to Viper
	viper.BindPFlags(pflag.CommandLine)

	// FIXME move other config here: slackToken, slackChannel, etc.
	viper.SetDefault("port", "8080")
	viper.SetDefault("data_dir", "data")
	viper.BindEnv("port", "PORT")
	viper.BindEnv("data_dir", "DATA_DIR")

	var bugout bool
	if value := os.Getenv("SLACK_TOKEN"); value == "" {
		fmt.Println("SLACK_TOKEN environment variable not set.")
		bugout = true
	}
	if value := os.Getenv("SLACK_CHANNEL"); value == "" {
		fmt.Println("SLACK_CHANNEL environment variable not set.")
		bugout = true
	}
	if bugout == true {
		os.Exit(1)
	}

	viper.MustBindEnv("slack_token", "SLACK_TOKEN")
	viper.MustBindEnv("slack_channel", "SLACK_CHANNEL")

	viper.SetDefault("discard_dir", viper.GetString("data_dir")+"/discard")
	viper.SetDefault("processed_dir", viper.GetString("data_dir")+"/processed")
	viper.SetDefault("uploads_dir", viper.GetString("data_dir")+"/uploads")
	viper.SetDefault("credentials_dir", viper.GetString("data_dir")+"/credentials")

	viper.BindEnv("data_dir", "DATA_DIR")
	viper.BindEnv("discard_dir", "DISCARD_DIR")
	viper.BindEnv("processed_dir", "PROCESSED_DIR")
	viper.BindEnv("uploads_dir", "UPLOADS_DIR")
	viper.BindEnv("credentials_dir", "CREDENTIALS_DIR")

	setupDirectory(viper.GetString("data_dir"))
	setupDirectory(viper.GetString("discard_dir"))
	setupDirectory(viper.GetString("processed_dir"))
	setupDirectory(viper.GetString("uploads_dir"))
	setupDirectory(viper.GetString("credentials_dir"))

}

func main() {
	pflag.CommandLine.Usage = func() {
		fmt.Fprintf(os.Stderr, "Custom help text goes here.\n\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", "your_app_name")
		pflag.PrintDefaults()
	}

	// Parse command line flags
	pflag.Parse()

	// Access the values using Viper
	configFile := viper.GetString("config")
	readySystemd := viper.GetBool("ready-systemd")

	// Load configuration file if specified
	if configFile != "" {
		viper.SetConfigFile(configFile)
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Println("Error reading config file:", err)
		}
	}
	// Check if --ready-systemd flag is provided
	if readySystemd {
		systemd_unit()
		os.Exit(0)
	}

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Use channels to coordinate goroutines
	done := make(chan struct{})

	go watchDirectory(viper.GetString("uploads_dir"), done)
	go receiver(done)

	// Block and wait for a signal to exit
	<-done

}
