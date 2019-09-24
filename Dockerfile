FROM golang:1.11.0-alpine3.8 as build

RUN apk --no-cache add git

COPY . /go/src/github.com/HenryTank/simpleP2P

#api
WORKDIR /go/src/github.com/HenryTank/simpleP2P
RUN go get && go install -v


FROM alpine:3.8

WORKDIR /
COPY --from=build /go/bin/simpleP2P .
#RUN apk --no-cache add ca-certificates openssl && update-ca-certificates

ARG COMMIT
ENV COMMIT ${COMMIT}

ENTRYPOINT ["/simpleP2P", "-port=23311"]
