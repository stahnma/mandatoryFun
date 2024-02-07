#!/bin/bash



tt() {
  curl -X POST -H "Content-Type: application/json" -d @./zoom/"$1" http://localhost:8889/
}

echo "Testing a join"
tt participant_joined.json
sleep 2

echo "Testing a leave"
tt participant_left.json
sleep 2

echo "Testing a join with a specific topic"
tt participant_joined_goodtopic.json
sleep 2

echo "Testing a join with a bad topic"
tt participant_special_meeting.json
sleep 2
