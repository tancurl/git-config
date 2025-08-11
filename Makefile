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

install: build
	@install -m 775 gcm /home/hos/.bin
