# hubot-slacklogs

Log messages in from slack for any channel hubot is in

# Installation

    npm i --save hubot-slacklogs

Edit your `external-scripts.json` file in your hubot applicaiton directory and add `hubot-slacklogs` to it.

# Usage

None, just invite bot to channels you wish to have logs for.

# Configuration

If `HUBOT_SLACK_LOGS_FILE` is set to a valid file path that hubot can write to,
it will logs jsonl format to that file. If that environment variable is not
set, it will log via `console.log` to stdout or whever the rest of your hubot
logs are going.

# License
MIT
