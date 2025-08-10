BINARY=gpgkeys
BIN_DIR=bin

all: build

tidy:
	@go mod tidy

build:
	@go build .

test:
	@go test ./...

clean:
	@rm -rf $(BIN_DIR)
