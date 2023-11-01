FROM golang:1.21.3-bullseye AS build

ENV GOCACHE=/go/.go/cache GOPATH=/go/.go/path TZ=Europe/London

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

FROM build AS test
