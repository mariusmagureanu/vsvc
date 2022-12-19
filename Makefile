export GO111MODULE=on
export GOARCH=amd64
export CGO_ENABLED=0

all: build

build:
	@go build -ldflags "-s -w" -o bin/vsvc

clean:
	@rm -rf ../bin || true
