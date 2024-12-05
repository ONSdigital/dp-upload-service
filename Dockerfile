FROM golang:1.23.2-bullseye as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

FROM build as test
