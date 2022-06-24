# build file
GO_BUILD=go build -ldflags -s -v
BINARY_NAME=das_database_server

# update
update:
	go mod tidy

# linux
parser_linux:
	export GOOS=linux
	export GOARCH=amd64
	$(GO_BUILD) -o $(BINARY_NAME) cmd/main.go
	mkdir -p bin/linux
	mv $(BINARY_NAME) bin/linux/
	@echo "build $(BINARY_NAME) successfully."

# mac
parser_mac:
	export GOOS=darwin
	export GOARCH=amd64
	$(GO_BUILD) -o $(BINARY_NAME) cmd/main.go
	mkdir -p bin/mac
	mv $(BINARY_NAME) bin/mac/
	@echo "build $(BINARY_NAME) successfully."

# win
parser_win: BINARY_NAME=das_database_server.exe
parser_win:
	export GOOS=windows
	export GOARCH=amd64
	$(GO_BUILD) -o $(BINARY_NAME) cmd/main.go
	mkdir -p bin/win
	mv $(BINARY_NAME) bin/win/
	@echo "build $(BINARY_NAME) successfully."

# default
default: parser_linux
