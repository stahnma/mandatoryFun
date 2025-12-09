# hubot-wisdom

Store and retrieve quotes of "wisdom" from the members of your chat.

# Installation

    npm i --save hubot-wisdom

Edit your `external-scripts.json` file in your hubot applicaiton directory and add `hubot-wisdom` to it.

# Usage

    > "Let It Be" -- Paul

    > hubot wisdom
    "A random quote is returned" -- how fun


# Configuration

## HUBOT_WISDOM_INCLUDE_TIMESTAMP

When set to `true`, the wisdom command will include the timestamp of when the quote was added.

- **Default**: `false`
- **Example**: `HUBOT_WISDOM_INCLUDE_TIMESTAMP=true`

When enabled, the output format will be:
```
"quote" -- author (added: YYYY-MM-DD HH:MM:SS)
```

# License
MIT
