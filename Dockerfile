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

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

RUN apt update \
    && apt install -y tzdata \
    && ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && dpkg-reconfigure --frontend noninteractive tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/das-database /app/das-database
COPY --from=build /app/config/config.yaml /app/config/config.yaml

EXPOSE 9090

ENTRYPOINT ["/app/das-database", "--config", "/app/config/config.yaml"]
