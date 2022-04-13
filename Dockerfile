# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.15-buster AS build

WORKDIR /app

COPY . ./

RUN go build -ldflags -s -v -o das-database cmd/main.go

##
## Deploy
##
FROM ubuntu

WORKDIR /app

COPY --from=build /app/das-database /app/das-database
COPY --from=build /app/config/config.yaml /app/config/config.yaml

EXPOSE 9090

ENTRYPOINT ["/app/das-database", "--config", "/app/config/config.yaml"]
