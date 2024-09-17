#!/usr/bin/env bash

# Development
URI=http://localhost:7171
# Produciton
URI=https://cspp.mandatoryfun.dev/upload


if [ -z "$1" ]; then
	echo "Usage: $0 <image file> <caption>"
	exit 1
fi
if [ -z "$2" ]; then
	echo "Usage: $0 <image file> <caption>"
	exit 1
fi
file="$1"
caption="$2"

curl -X POST \
  -F "image=@$file" \
  -F "caption=$caption"  \
  -H "X-API-KEY: $API_KEY" \
$URI

