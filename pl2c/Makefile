

#GOCMD=go
#GOBUILD=$(GOCMD) build
#GOCLEAN=$(GOCMD) build
ASSETS=assets.go

all: build

build:
	goimports -w -l .
	goblet templates > $(ASSETS)
	go build

clean:
	go clean
	rm -f $(ASSETS)

deps:
	go get github.com/deadcheat/goblet

