FROM golang:1.22-bullseye as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

FROM build as test
