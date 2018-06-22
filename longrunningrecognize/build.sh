#!/usr/bin/env bash

PROJECT_DIR=${PWD}
DIST_DIR="${PROJECT_DIR}/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
BOLD='\033[1m'
NA='\033[0m'

#-------------------------------------------------------------------------------------------------
# Setup environment
#-------------------------------------------------------------------------------------------------
export GOPATH="${PROJECT_DIR}:${HOME}/go"
export GOOS=linux
export GOARCH=amd64

#-------------------------------------------------------------------------------------------------
# Compile sources
#-------------------------------------------------------------------------------------------------
echo -e "${GREEN}${BOLD}Compiling sources...${NC}${NA}"

/usr/local/go/bin/go build -buildmode=exe -ldflags="-v -s" -o ${DIST_DIR}/longrunningrecognize ${PROJECT_DIR}/*.go

if [ $? -ne 0 ]
then
	echo -e "${RED}${BOLD}Failed to compile sources, interrupted.${NC}${NA}"
	exit 1
fi

echo -e "${GREEN}${BOLD}Completed.${NC}${NA}"

#-------------------------------------------------------------------------------------------------
# Copy dependencies
#-------------------------------------------------------------------------------------------------
echo -e "${GREEN}${BOLD}Copying dependencies...${NC}${NA}"

cp -afv --recursive --copy-contents --target-directory=${DIST_DIR} ${PROJECT_DIR}/../voices/*

if [ $? -ne 0 ]
then
	echo -e "${RED}${BOLD}Failed to copy dependencies, interrupted.${NC}${NA}"
	exit 1
fi

echo -e "${GREEN}${BOLD}Completed.${NC}${NA}"
