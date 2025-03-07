FROM golang:1.24.1 as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

FROM build as test
