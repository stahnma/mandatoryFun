// Description:
//   Request an API key from CSPP via slack interacions
//
// Dependencies:
//   None
//
// Configuration:
//   HUBOT_CSPP_API_URI - Defaults to localhost:7177
//
// Commands:
//   hubot api key - request an API Key from CSPP
//
// Author:
//   stahnma
//
// Category: social
//
module.exports = function (robot) {
  var env = process.env;
  var uri = env.HUBOT_CSPP_API_URI;
  if (uri == undefined || uri == "") {
    uri = "http://localhost:7177/api";
  }

  robot.respond(/cspp api key/i, function (msg) {
    const slackId = msg.message.user.id;
    const requestData = { slack_id: slackId };

    robot
      .http(uri)
      .header("Content-Type", "application/json")
      .post(JSON.stringify(requestData))((err, res, body) => {
      if (err) {
        robot.logger.debug(err);
        msg.send(err);
        return;
      }

      if (res.statusCode === 200) {
        msg.send("Request successful!, check your DMs");
      } else {
        msg.send("Request failed with status code: " + res.statusCode);
      }
    });
  });
};
