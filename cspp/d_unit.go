package main

import (
	"fmt"
	"os"
	"os/user"
	"text/template"
)

const systemdUnitTemplate = `
[Unit]
Description={{.Description}}
After=network.target

[Service]
ExecStart={{.ExecStart}}
Restart=always
User={{.User}}

[Install]
WantedBy=multi-user.target
`

// verify the envionment variables are set
// FIXME move all of this to viper
func verifyEnvVars() {
	requiredEnvVars := []string{"SLACK_TOKEN", "IMAGE_DIRECTORY", "SLACK_CHANNEL"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			fmt.Println("Missing required environment variable: " + envVar)
			os.Exit(1)
		}
	}
}

// SystemdUnit represents the data for the systemd unit template
type SystemdUnit struct {
	Description string
	ExecStart   string
	User        string
}

func systemd_unit() {

	verifyEnvVars()
	// Define the data for the systemd unit template
	fq_program, err := os.Executable()
	if err != nil {
		panic(err)
	}
	current_user, _ := user.Current()

	unitData := SystemdUnit{
		Description: os.Args[0],
		ExecStart:   fq_program,
		User:        current_user.Username,
	}

	// Create a new template and parse the systemd unit template string
	tmpl, err := template.New("systemdUnit").Parse(systemdUnitTemplate)
	if err != nil {
		panic(err)
	}

	// Execute the template with the provided data and write to stdout
	err = tmpl.Execute(os.Stdout, unitData)
	if err != nil {
		panic(err)
	}
}
