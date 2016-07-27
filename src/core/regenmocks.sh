#!/usr/bin/env bash

export GOPATH=$PWD/../../vendor:$PWD/../../
mockgen -package=fixtures core/models CardRepository > fixtures/mock_shopcard.go
