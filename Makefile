.PHONY: all build

GO=go
BIN=bin/searchxml

default: all

all: bin/searchxml

bin/searchxml:
	$(GO) build -o ${BIN} main.go
