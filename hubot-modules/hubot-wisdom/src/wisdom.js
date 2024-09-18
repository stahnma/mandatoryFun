// Description:
//   Store and retrieve quotes along with the author, submitter, and timestamp.
//
// Commands:
//   hubot "<quote>" -- <author> - Store a new quote along with your username and the current time.
//   hubot wisdom - Responds with a random quote from the memory.

// Author: stahnma

// Category: social

const {
  WebClient
} = require('@slack/web-api');

module.exports = (robot) => {

  // Listening for quotes and storing them
  robot.hear(/^\s*(\"|“)(.+?)(\"|”)\s+(--|—)\s*(.+?)$/, async (msg) => {
    // Extract quote and author from the message
    const quote = msg.match[2];
    const author = msg.match[5];
    const user = msg.message.user.name; // Capturing the user who added the quote
    const timestamp = new Date().toISOString().replace(/T/, ' ').replace(/\..+/, ''); // ISO format to YYYY-MM-DD HH:MM:SS

    // Initialize the quotes array in the brain if it doesn't exist
    if(!robot.brain.data.quotes) {
      robot.brain.data.quotes = [];
    }

    // Store the new quote with user and timestamp
    robot.brain.data.quotes.push({
      quote: `"${quote}"`,
      author,
      user,
      timestamp
    });

    // Check if the bot is running in Slack
    if(robot.adapterName === 'slack') {
      const slackMessage = msg.message.rawMessage;

      if(slackMessage && slackMessage.ts && slackMessage.channel) {
        try {
          // Create a new instance of WebClient using the bot's token
          const token = process.env.HUBOT_SLACK_TOKEN || process.env.SLACK_BOT_TOKEN; // Ensure the token is set in environment variables
          const web = new WebClient(token);

          // Add a reaction to the message
          await web.reactions.add({
            name: 'quote',
            channel: slackMessage.channel,
            timestamp: slackMessage.ts
          });

          robot.logger.info('Reaction added successfully');
        } catch (err) {
          robot.logger.error('Failed to add reaction:', err);
        }
      } else {
        robot.logger.error('Missing Slack message data: ts or channel');
      }
    } else {
      // Fallback if not running in Slack
      msg.send(`Quote added.`);
    }
  });

  // Responding with a random quote
  robot.respond(/wisdom$/i, (msg) => {
    const quotes = robot.brain.data.quotes;

    // Check if there are any quotes stored
    if(quotes && quotes.length > 0) {
      const randomIndex = Math.floor(Math.random() * quotes.length);
      const {
        quote,
        author
      } = quotes[randomIndex];
      msg.send(`${quote} -- ${author}`);
    } else {
      msg.send("I have no wisdom to share yet. Please teach me.");
    }
  });
};
