#!/usr/bin/env bash

#-------------------------------------------------------------------------------------------------
# Load voice data as a base64 encoded string
#-------------------------------------------------------------------------------------------------
voice_data=$(base64 --wrap=0 "$(dirname "${0}")/../voices/1min.flac")

#-------------------------------------------------------------------------------------------------
# Compose request body
#-------------------------------------------------------------------------------------------------
request_body="{ \"signal\": \"${voice_data}\", \"language_code\": \"en-US\", \"execute_beam_search\": false }"

#-------------------------------------------------------------------------------------------------
# Execute remote call
#-------------------------------------------------------------------------------------------------
echo -En "${request_body}" | curl \
	--verbose \
	--max-time 60 \
	--request POST \
	--header "Content-Type: application/json; charset=utf-8" \
	--data-binary @- \
	"https://asr.sapiensapi.com/v1/speech:recognize"
