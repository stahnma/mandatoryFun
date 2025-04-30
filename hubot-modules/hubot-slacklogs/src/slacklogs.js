// Description:
//  Log your slack messages in jsonl format
//
// Configuration:
//    HUBOT_SLACK_LOGS_FILE - absolute path to file where logs should be placed
//    HUBOT_SLACK_TOKEN - needed to use slack API for room lookups, etc
//
// Author: stahnma
//
// Category: workflow
const fs = require('fs');
const path = require('path');
const { WebClient } = require('@slack/web-api');

// Config
const logFilePath = process.env.HUBOT_SLACK_LOGS_FILE;
const slackToken = process.env.HUBOT_SLACK_TOKEN;

let logStream = null;
if (logFilePath) {
  try {
    const dir = path.dirname(logFilePath);
    fs.mkdirSync(dir, { recursive: true });
    logStream = fs.createWriteStream(logFilePath, { flags: 'a' });
    console.log(`[hubot-logger] Logging to file: ${logFilePath}`);
  } catch (err) {
    console.error(`[hubot-logger] Failed to set up log file: ${err.message}`);
  }
}

const slackClient = slackToken ? new WebClient(slackToken) : null;

// In-memory cache for room name/type
const roomCache = new Map();

function getSlackRoomType(roomId) {
  if (!roomId || typeof roomId !== 'string') return 'unknown';
  if (roomId.startsWith('C')) return 'public_channel';
  if (roomId.startsWith('G')) return 'private_channel';
  if (roomId.startsWith('D')) return 'direct_message';
  if (roomId.startsWith('Q')) return 'group_dm';
  if (roomId.startsWith('T')) return 'external_dm';
  return 'unknown';
}

async function getRoomInfo(roomId) {
  const cached = roomCache.get(roomId);
  const now = Date.now();

  if (cached && now - cached.fetchedAt < 60 * 60 * 1000) {
    return cached.info;
  }

  if (!slackClient || !roomId) {
    return { name: null, type: getSlackRoomType(roomId) };
  }

  try {
    const result = await slackClient.conversations.info({ channel: roomId });
    const roomName = result.channel?.name || null;
    const roomType = result.channel?.is_im
      ? 'direct_message'
      : result.channel?.is_private
      ? 'private_channel'
      : 'public_channel';

    const info = { name: roomName, type: roomType };
    roomCache.set(roomId, { info, fetchedAt: now });
    return info;
  } catch (err) {
    console.error(`[hubot-logger] Failed to fetch room info for ${roomId}: ${err.message}`);
    return { name: null, type: getSlackRoomType(roomId) };
  }
}

module.exports = (robot) => {
  if (robot.adapterName !== 'slack') {
    console.log(`[hubot-logger] Adapter is '${robot.adapterName}', skipping Slack-specific logging.`);
    return;
  }

  robot.hear(/.*/, async (res) => {
    const rawMessage = res?.message?.rawMessage;
    const roomId = res.message.room;

    if (!rawMessage || typeof rawMessage !== 'object' || !roomId) return;

    try {
      const roomInfo = await getRoomInfo(roomId);

      const message = {
        user: res.message.user.name,
        userId: res.message.user.id,
        text: res.message.text,
        roomId: roomId,
        roomName: roomInfo.name,
        roomType: roomInfo.type,
        timestamp: new Date().toISOString(),
        slackTimestamp: rawMessage.ts || null,
        threadTimestamp: rawMessage.thread_ts || null,
        isThreadRoot: rawMessage.thread_ts ? rawMessage.thread_ts === rawMessage.ts : false,
      };

      const line = JSON.stringify(message);

      if (logStream) {
        logStream.write(line + '\n');
      } else {
        console.log(line);
      }
    } catch (err) {
      console.error('[hubot-logger] Error while logging message:', err);
    }
  });
};

