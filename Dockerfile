# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18.10-buster AS build

WORKDIR /app

COPY . ./

ENV GOPROXY=https://goproxy.cn,direct

RUN go build -ldflags -s -v -o das-database cmd/main.go

##
## Deploy
##
FROM ubuntu

ARG TZ=Asia/Shanghai

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt-get update \
    && apt-get install -y ca-certificates tzdata \
    && ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && dpkg-reconfigure tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /app/das-database /app/das-database
COPY --from=build /app/config/config.example.yaml /app/config/config.yaml

EXPOSE 8118

ENTRYPOINT ["/app/das-database", "--config", "/app/config/config.yaml"]
