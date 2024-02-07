# What is this?

This is a small program designed to run on a public network that can receive webhooks from zoom for participants joining/leaving a meeting. It then translates those webhooks into webhook calls to slack to notify a specific slack channel about a participant joining or leaving. This was designed to work with long-standing/running meetings.

It also can notify IRC channels using similar methodology to the slack send, however instead of using a webhook, it just speaks native IRC protocols.

## Assumptions

1. You have a long-standing zoom meeting that has a constant ID. I use a paid account, I'm not sure if you have to have one or not to do this.
1. You have access/permissions to create a zoom application.
1. You have access/permissions to create a slack webhook integration.
1. You optionally have an IRC server.


# Running Config

you must set

`ZOOM_SECRET` which is the `secret token` provided on your zoom application. This is how zoom knows your webhook listener is the intended target.

`ZOOMWH_PORT` the port to put this webhook listener on. Default is 8888.

`ZOOMWH_MSG_SUFFIX` the name of the call after "Person has <left|joined>" it defaults to "the zoom meeting".

`ZOOMWH_MEETING_NAME` only notify on a particular meeting name. Default is any/all. Filter is a literal string, not a regex.

## Slack Messaging
`ZOOMWH_SLACK_ENABLE` should be set to 'true' if using this feature. It defaults to true.

`ZOOMWH_SLACK_WH_URI` is the slack webhook uri. This can be a comma separated list of URIs. Required if `ZOOMWH_SLACK_ENABLE` is `true`. No Default.

## IRC Messaging
`ZOOMWH_IRC_ENABLE` true will enable this. Default is false.

`ZOOMWH_IRC_SERVER` is the a URI and includes a port. Required. No Default.

`ZOOMWH_IRC_NICK` the nick name to use when posting for IRC and authenticating against the server. Required. No Default.

`ZOOMWH_IRC_AUTH_PASS` is the password to authenticate with for the IRC Server.

`ZOOMWH_IRC_CHANNEL` is where to post the messages. There is not default. This is required if `ZOOMWH_IRC_ENABLE` is true.

# Building

Checkout the source code

    go mod tidy
    go build .

# Running via systemd

There is a unit file in the `contrib` directory. It should get you most of the way there.

Basically, create a user by doing something like `useradd zoomwh`.

Then copy the unit file in `/usr/lib/systemd/system`.

Make a file with your environment settings and place it in `/etc/sysconfig` or whereever your distribution does that stuff.

`systemctl enable zoomwh.service`

`systemctl start zoomwh.service`


# Limitations
1. This can currently be set up for one zoom meeting (because there is no logic looking what the meeting is called or anything).
1. You can't make the POST URL path anything other than "/" right now. Again, this is an easy fix, but I haven't done it yet. Ideally you could pass in a ENV var for that.


# Setup

## Assumptions

1. You have a long-standing zoom meeting that has a constant ID. I use a paid account, I'm not sure if you have to have one or not to do this.
1. You have access/permissions to create a zoom application.
1. You have access/permissions to create a slack webhook integration.
1. You are comfortable putting a service on a public network. (I front mine with a reverse proxy).
1. You are an IRC admin if you're using the IRC functionality.

## Zoom

See [Zoom Documenation](https://github.com/stahnma/zoomwh/blob/main/docs/zoom_app_creation.md)

## Slack

See [Slack App Documentation](https://github.com/stahnma/zoomwh/blob/main/docs/slack_integrations.md)

# LICENSE
MIT