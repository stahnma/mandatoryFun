#!/usr/bin/env bash

# Development
URI=http://localhost:7171
# Produciton
URI=https://cspp.mandatoryfun.dev/upload

curl -X DELETE -H "X-API-KEY: $API_KEY" $URI

