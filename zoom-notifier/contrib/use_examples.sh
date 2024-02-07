#!/bin/bash

if [ -z "$1" ] ; then
	echo "Usage $0 json_file"
	exit 1
fi

curl -X POST -H "Content-Type: application/json" -d @"$1" http://localhost:8888/
