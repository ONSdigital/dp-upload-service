FROM golang:1.18-stretch as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

FROM build as test
