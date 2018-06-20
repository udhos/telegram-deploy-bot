#!/bin/bash

go get gopkg.in/telegram-bot-api.v4

build() {
	local i="$1"
	gofmt -s -w "$i"
	go tool fix "$i"
	go tool vet "$i"

	hash 2>/dev/null golint && golint "$i"

	go test "$i"
	go install "$i"
}

build ./telegram-bot-send
build ./telegram-deploy-bot

