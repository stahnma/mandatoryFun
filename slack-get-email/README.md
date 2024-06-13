# Slack Channel Members Email Extractor

This Go program retrieves and prints the email addresses of all members in a
specified Slack channel. This could be useful to send meeting invitations and
plan those fun times with yhour channel.

## Prerequisites

- Go 1.16 or later
- A Slack App with the following OAuth scopes:
  - `channels:read`
  - `groups:read`
  - `users:read`
  - `users:read.email`
  - `conversations.members`

## Setup

1. **Create a Slack App:**
   - Go to the [Slack API: Applications](https://api.slack.com/apps) page.
   - Click "Create New App" and follow the prompts to create your app.

2. **Configure OAuth Scopes:**
   - Go to the "OAuth & Permissions" page for your app.
   - Under "OAuth Scopes", add the following scopes:
     - `channels:read`
     - `groups:read`
     - `users:read`
     - `users:read.email`
     - `conversations.members`
   - Save the changes.

3. **Install the App to your Workspace:**
   - After adding the scopes, you will see an "Install App to Workspace" button. Click it and follow the prompts to install the app.
   - Once installed, you'll receive an OAuth token that includes the necessary scopes.

4. **Set Environment Variables:**
   - Set the OAuth token and the channel ID as environment variables.

   ```sh
   export SLACK_TOKEN="your-oauth-access-token"
   export SLACK_CHANNEL="your-channel-id"
   ```

## Usage

1. **Clone the Repository:**

   ```sh
   git clone https://github.com/yourusername/slack-email-extractor.git
   cd slack-email-extractor
   ```

2. **Run the Program:**

   ```sh
   go run main.go
   ```

3. **Output:**

   The program will print the real names and email addresses of all members in the specified Slack channel.

## Example Output

```
John Doe <john.doe@example.com>
Jane Smith <jane.smith@example.com>
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

