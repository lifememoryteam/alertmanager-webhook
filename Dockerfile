FROM golang:1.12.10

MAINTAINER Shogo Maeda

ENV GO111MODULE=on

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alertmanager-webhook

EXPOSE 8000

CMD [ "./alertmanager-webhook" ]
