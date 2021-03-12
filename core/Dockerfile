FROM golang:alpine

WORKDIR /app

COPY ./src /app

RUN apk add --no-cache gcc musl-dev

RUN go mod download

RUN go test ./... -v

RUN go install ./cmd/jt-slackbot-core

ENTRYPOINT ["/go/bin/jt-slackbot-core"]