# hubot-message-aggregator

![95feb390-e7b4-4033-b02a-49b1557efd6f](https://github.com/stahnma/mandatoryFun/assets/6961/84f209cb-8ad4-41ac-ac50-c41018d71f28)


This is a simplistic module designed to aggregate message into a single channel
based upon an emoji reaction.

The idea is that if a people react to a message, it will "bookmark" it, or
aggregate it into a channel. An example could be to have any message where
people react to a message with a `:thankyou:` emoji  placed into a `#thanks` or
`#gratitude` channel.

Another example could be that any messages with a :computer: emoji reaction
gets passed to an incident or alert channel.


# Installation

    npm i --save hubot-message-aggregator


Edit your `external-scripts.json` file in your hubot application directory and
add `hubot-post-aggregator` to it.

# Configuration 

## Variables

By default, the pattern it looks for in the emoji name is "thank". This can be
set via `HUBOT_AGGREGATION_PATTERN`. The value for it is just the string portion
of the regex (e.g. no need for `/` around it.)

By default, the aggregator will *not* post messages from private channels or
conversations. This can be overridden by setting
`HUBOT_AGGREGATION_FROM_PRIVATE_CONVERSATIONS` to `true`.

`HUBOT_AGGREGATION_CHANNEL` is required. It can be specified as a Slack
ChannelID such as `C1234567890` or a string such as `general`. 

:warning: Note, the `#` should not be included.

## Behavior

To reduce noise, in case a message gets several reactions that match the
pattern, a message is only posted to the aggregation target channel once per
24 hours. 

The aggregator does aggregate or repost messages inside the aggregator channel, as
this gets into a :turtles: all the way down type situation.

Ouput and information about how this is running can be found in the hubot log.

### Limitations

Right now, the 24 hours period is not configurable. Perhaps it should be.

There is exactly 1 ruleset. It might be nice to have several options for this
like `:sirens:` go the `incident` channel and `:prayinghands:` go to a `thanks` channel.
Patches welcome.


# License
MIT
