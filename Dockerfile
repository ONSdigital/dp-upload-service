FROM golang:1.16-stretch as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

FROM build as test
