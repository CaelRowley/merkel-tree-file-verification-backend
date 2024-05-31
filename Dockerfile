FROM golang:1.22.1

WORKDIR /app

COPY . .

RUN go mod download && go mod verify

WORKDIR /app/cmd

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
