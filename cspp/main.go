package main

import (
	"fmt"
	"net/url"
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

	viper.SetDefault("port", "8080")
	viper.SetDefault("data_dir", "data")
	viper.SetDefault("base_url", "")
	viper.BindEnv("port", "CSPP_PORT")
	viper.BindEnv("data_dir", "CSPP_DATA_DIR")

	var bugout bool
	if value := os.Getenv("CSPP_SLACK_TOKEN"); value == "" {
		fmt.Println("CSPP_SLACK_TOKEN environment variable not set.")
		bugout = true
	}
	if value := os.Getenv("CSPP_SLACK_CHANNEL"); value == "" {
		fmt.Println("CSPP_SLACK_CHANNEL environment variable not set.")
		bugout = true
	}
	if value := os.Getenv("CSPP_SLACK_TEAM_ID"); value == "" {
		fmt.Println("CSPP_SLACK_TEAM_ID environment variable not set.")
		bugout = true
	}
	if bugout == true {
		os.Exit(1)
	}

	viper.MustBindEnv("slack_token", "CSPP_SLACK_TOKEN")
	viper.MustBindEnv("slack_channel", "CSPP_SLACK_CHANNEL")
	viper.MustBindEnv("slack_team_id", "CSPP_SLACK_TEAM_ID")

	viper.SetDefault("discard_dir", viper.GetString("data_dir")+"/discard")
	viper.SetDefault("processed_dir", viper.GetString("data_dir")+"/processed")
	viper.SetDefault("uploads_dir", viper.GetString("data_dir")+"/uploads")
	viper.SetDefault("credentials_dir", viper.GetString("data_dir")+"/credentials")

	viper.BindEnv("data_dir", "CSPP_DATA_DIR")
	viper.BindEnv("discard_dir", "CSPP_DISCARD_DIR")
	viper.BindEnv("processed_dir", "CSPP_PROCESSED_DIR")
	viper.BindEnv("uploads_dir", "CSPP_UPLOADS_DIR")
	viper.BindEnv("credentials_dir", "CSPP_CREDENTIALS_DIR")
	viper.BindEnv("base_url", "CSPP_BASE_URL")

	setupDirectory(viper.GetString("data_dir"))
	setupDirectory(viper.GetString("discard_dir"))
	setupDirectory(viper.GetString("processed_dir"))
	setupDirectory(viper.GetString("uploads_dir"))
	setupDirectory(viper.GetString("credentials_dir"))

	validatePortVsBaseURL()

}

func validatePortVsBaseURL() {
	log.Debugln("validatePortVsBaseURL")
	baseurl := viper.GetString("base_url")
	port := viper.GetString("port")
	if baseurl != "" && port != "" {
		parsedURL, err := url.Parse(baseurl)
		if err != nil {
			log.Errorln("Error parsing base URL:", err)
			os.Exit(1)
		}
		baseport := parsedURL.Port()
		if baseport == "" && port != "" {
			return
		}
		if baseport != port {
			viper.Set("port", baseport)
			if port != "8080" {
				log.Infoln("CSPP_PORT overridden by value specified in CSPP_BASE_URL.")
			}
		}
	}
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
