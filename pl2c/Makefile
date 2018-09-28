

#GOCMD=go
#GOBUILD=$(GOCMD) build
#GOCLEAN=$(GOCMD) build

all: build

build:
	goimports -w -l .
	statik -src templates -m -f
	go build

test: build
	go test

clean:
	go clean
	rm -rf statik

deps:
	go get github.com/rakyll/statik

