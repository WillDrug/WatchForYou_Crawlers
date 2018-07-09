FROM golang:1.8

LABEL maintainer="WillDrug"

WORKDIR /go/src/youtube_crawler
COPY . .

CMD "go" "run" "main.go"
