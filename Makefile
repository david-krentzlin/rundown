SHELL := /usr/bin/env bash

APP_NAME := rundown
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)
GO ?= go

.PHONY: help build run run-demo test quality-gate clean

help:
	@printf '%s\n' \
		'Targets:' \
		'  make build         Build the app binary to ./$(BIN)' \
		'  make run           Run the app with default content' \
		'  make run-demo      Run the app with examples/scroll-demo.md' \
		'  make test          Run all tests' \
		'  make quality-gate  Run deterministic checks via agent_ops' \
		'  make clean         Remove built artifacts'

build:
	@mkdir -p "$(BIN_DIR)"
	@$(GO) build -o "$(BIN)" ./cmd/rundown

run:
	@$(GO) run ./cmd/rundown

run-demo:
	@$(GO) run ./cmd/rundown examples/scroll-demo.md

test:
	@$(GO) test ./...

quality-gate:
	@agent_ops make quality-gate

clean:
	@rm -rf "$(BIN_DIR)"
