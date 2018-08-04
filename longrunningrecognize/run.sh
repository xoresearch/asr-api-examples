#!/usr/bin/env bash

PROJECT_DIR="$(dirname "${0}")"

API_ENDPOINT="https://asr.sapiensapi.com"
VOICE_RECORD="${PROJECT_DIR}/../voices/20min.flac"

#-------------------------------------------------------------------------------------------------
# Setup environment
#-------------------------------------------------------------------------------------------------
export PATH="${PATH}:/usr/local/go/bin"
export GOPATH="${HOME}/go"
export GOOS=linux
export GOARCH=amd64

#-------------------------------------------------------------------------------------------------
# Run the test
#-------------------------------------------------------------------------------------------------
go run ${PROJECT_DIR}/*.go ${API_ENDPOINT} ${VOICE_RECORD}
