package main

import (
	"fmt"
	"github.com/slack-go/slack"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL")

	if token == "" || channelID == "" {
		log.Fatal("SLACK_TOKEN (xoxb-...) and SLACK_CHANNEL (channel id) environment variables are required")
	}

	api := slack.New(token)

	// Get all members of the channel
	users, err := getChannelMembers(api, channelID)
	if err != nil {
		log.Fatalf("Failed to get channel members: %s", err)
	}
	// Sort users alphabetically by RealName, case-insensitive
	sort.Slice(users, func(i, j int) bool {
		return strings.ToLower(users[i].Profile.RealName) < strings.ToLower(users[j].Profile.RealName)
	})

	// Print email addresses of the members
	for _, user := range users {
		if user.Profile.Email == "" {
			continue
		}
		fmt.Println(user.Profile.RealName + " <" + user.Profile.Email + ">")
	}
}

// getChannelMembers retrieves all users in a Slack channel
func getChannelMembers(api *slack.Client, channelID string) ([]slack.User, error) {
	var allUsers []slack.User
	cursor := ""

	// Retrieve user IDs in the channel
	for {
		members, nextCursor, err := api.GetUsersInConversation(&slack.GetUsersInConversationParameters{
			ChannelID: channelID,
			Cursor:    cursor,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get users in conversation: %w", err)
		}

		// Retrieve user details for each member
		for _, memberID := range members {
			user, err := api.GetUserInfo(memberID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user info for %s: %w", memberID, err)
			}
			allUsers = append(allUsers, *user)
		}

		// If there's no next cursor, break the loop
		if nextCursor == "" {
			break
		}

		cursor = nextCursor
	}

	return allUsers, nil
}
