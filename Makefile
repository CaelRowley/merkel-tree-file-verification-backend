.PHONY: all clean build test run
BINARY_NAME=main

all: clean build test

clean:
	rm -f ${BINARY_NAME}

build:
	go build -o ${BINARY_NAME} cmd/main.go

test:
	go test ./...

run: build
	docker-compose -f docker-compose.yml up backend 
	./${BINARY_NAME}
