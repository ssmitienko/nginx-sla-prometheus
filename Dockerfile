##
## Build
##
FROM golang:1.18-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN CGO_ENABLED=0 go build -o /nginx-sla-prometheus

##
## Deploy
##
FROM alpine/curl
WORKDIR /

COPY --from=build /nginx-sla-prometheus /nginx-sla-prometheus

EXPOSE 9009

USER nobody:nobody

ENTRYPOINT ["/nginx-sla-prometheus", "-listen" , "0.0.0.0:9009" , "-user", "sla", "-password", "password", "-backend", "http://localhost/sla/"]