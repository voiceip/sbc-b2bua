#!/bin/bash

GOOS=linux go build -gcflags "all=-N -l" -o sbc-b2bua.linux 