package main

import (
	"crypto/tls"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/thoj/go-ircevent"
)

func sendIRC(message string) {
	log.Debugln("(sendIRC) Sending IRC notification", message)
	ircEnable := viper.GetString("irc_enable")
	var ircServer, ircChannel, ircNick, ircPassword string
	if ircEnable == "true" {
		ircServer = viper.GetString("irc_server")
		ircChannel = viper.GetString("irc_channel")
		ircNick = viper.GetString("irc_nick")
		ircPassword = viper.GetString("irc_pass")
	}

	irccon := irc.IRC(ircNick, ircNick)
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.Password = ircPassword
	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(ircChannel)
	})

	err := irccon.Connect(ircServer)
	if err != nil {
		log.Errorln("Error connecting to IRC server:", err)
	}
	defer irccon.Quit()

	irccon.Privmsg(ircChannel, message)

	time.AfterFunc(1*time.Second, func() {
		irccon.Quit()
	})
	irccon.Loop()
}
