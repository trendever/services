#!/usr/bin/env bash

export GOPATH=$PWD/../../vendor:$PWD/../../
mockgen -package=fixtures chat/models ConversationRepository > fixtures/mock_conversation.go
