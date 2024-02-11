# Collaborative Sh💩t Posting Pipeline (CSPP)

## Use Case

Have you been making stupid/silly/wonderful AI pictures and art work? Do you want to share it with your coworkers and friends? If yes, this is the simple program for you.

![Slack_-_ai-junk_-_stahnma](https://github.com/stahnma/mandatoryFun/assets/6961/f7f25bb2-33b2-40f8-bb62-b87d49904f62)


## Overview
This program has two parts. The first part is a simple web server that listens for POST requests. When it receives a POST request, it saves the image to a directory. The second part is a program that reads from a directory and sends all images found to a slack channel.

It's a queue for posting images to slack with API security based on being a member of a slack team.

## Client Side

There is a hubot plugin that can be used to request an API key. See [hubot-cspp-client](../hubot-modules/hubot-cspp-client/README.md) for more information.

## Configuration

To run this you simple start the binary after having the correct environment variables set. Environment variables are as follows:

You need to set the following enviornment variables.
| Variable Name       | Description                                            | Required | Example                                                      | Default Value                   |
|---------------------|--------------------------------------------------------|----------|--------------------------------------------------------------|---------------------------------|
| `CSPP_SLACK_TOKEN`    | slack token used to send message to slack              | required | `xoxb-1234567890-1234567890123-12345678901234567890abcdef123456` | none                            |
| `CSPP_DATA_DIR`       | directory to save images, credentials, invalid uploads, images sent, etc | required | `/var/lib/cspp`                                              | `./data`                        |
| `CSPP_SLACK_CHANNEL` | slack channel to send images to                        | required | `#cspp`                                                      | none                            |
| `CSPP_SLACK_TEAM_ID`  | slack team id to validate requests                     | required | `T12345678`                                                  | none                            |
| `CSPP_PORT`           | port to listen on                                      | optional | `8080`                                                       | `8080`                          |
| `CSPP_BASE_URL`       | base url for the server                                | optional | `https://cspp.example.com`                                   | `.` (meaning all paths are relative) |
| `CSPP_DISCARD_DIR`    | directory to save invalid uploads                      | optional | `/var/lib/cspp/discard`                                      | `./data/discard`                |
| `CSPP_PROCESSED_DIR`  | directory to save images sent                          | optional | `/var/lib/cspp/processed`                                    | `./data/processed`              |
| `CSPP_UPLOADS_DIR`    | directory to save images                               | optional | `/var/lib/cspp/uploads`                                      | `./data/uploads`                |
| `CSPP_CREDENTIALS_DIR`| directory to save API keys as json blobs               | optional | `/var/lib/cspp/credentials`                                  | `./data/credentials`           |

:warning: If you specifiy `CSPP_BASE_URL` with a port on the string and specify `CSPP_PORT` and they do not match, you may get unpredictable results

## Slack Specifics

### Finding the Team ID

The easiest way to find the Slack Team ID is to open the slack team in a web browser and look at the URL. Once you open the team on the web (you may have to manually click "open in browser" so that it doesn't jump into your desktop application), you can look at the URL. The team ID is the string that start with `T` after the `client/` and looks something like `T047M58T6`.

### Finding your Slack User ID

To find your SLack User ID (only needed if you're not using the [hubot cspp client](../hubot-modules/hubot-cspp-client)) you can click on your profile in the slack app and then click "More" and then "Copy Member ID".

## Proxying

CSPP listens in clear text. If you want SSL (and you should), you should put a reverse proxy in front of it. We recommend using [Caddy](https://caddyserver.com/) for this purpose. Here is an example Caddyfile:

```caddy
cspp.example.com {

    reverse_proxy localhost:8080

    log {
        output file /var/log/caddy/cspp/caddy.log
        level info
    }
}
```
## Contributions

We use a [Conventional Commit](https://www.conventionalcommits.org/en/v1.0.0/) style for commit messages and this is enforced by CI.

## License
MIT

© Michael Stahnke 2023,2024
