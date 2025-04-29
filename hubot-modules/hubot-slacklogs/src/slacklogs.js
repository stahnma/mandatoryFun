// Description:
//   Store and retrieve quotes along with the author, submitter, and timestamp.
//
// Commands:
//   hubot "<quote>" -- <author> - Store a new quote along with your username.
//   hubot wisdom - Responds with a random quote from the wisdom bank..
//
// Author: stahnma
//
// Category: social

const {
  WebClient
} = require('@slack/web-api');

const fs = require('fs');
const path = require('path');

// Capture env var
const logFilePath = process.env.HUBOT_SLACK_LOGS_FILE;
let logStream = null;

// Prepare write stream if logging to file
if (logFilePath) {
  try {
    // Ensure directory exists
    const dir = path.dirname(logFilePath);
    fs.mkdirSync(dir, { recursive: true });

    logStream = fs.createWriteStream(logFilePath, { flags: 'a' });
    console.log(`[hubot-logger] Logging to file: ${logFilePath}`);
  } catch (err) {
    console.error(`[hubot-logger] Failed to set up log file: ${err.message}`);
  }
}

module.exports = (robot) => {
  robot.hear(/.*/, async (res) => {
    const rawMessage = res.message.rawMessage;

    const message = {
      user: res.message.user.name,
      userId: res.message.user.id,
      text: res.message.text,
      room: res.message.room,
      timestamp: new Date().toISOString(),
      slackTimestamp: rawMessage.ts,
      threadTimestamp: rawMessage.thread_ts || null,
      isThreadRoot: rawMessage.thread_ts ? rawMessage.thread_ts === rawMessage.ts : false,
    };

    const line = JSON.stringify(message);

    if (logStream) {
      logStream.write(line + '\n');
    } else {
      console.log(line);
    }
  });
};

