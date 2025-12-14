.PHONY: swag format

SWAG_BIN=~/go/bin/swag
MAIN_FILE=cmd/api/main.go
OUTPUT_DIR=./api/docs

swag:
	$(SWAG_BIN) init -g $(MAIN_FILE) --parseDependency --parseInternal --parseVendor -o $(OUTPUT_DIR)

format:
	go fmt ./...