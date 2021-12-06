# syntax = docker/dockerfile:experimental
FROM golang:1.17.0 as builder

ENV GO111MODULE on
ENV GOPRIVATE "bitbucket.org/latonaio"
WORKDIR /go/src/latonaio
COPY go.mod .


RUN git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"
RUN mkdir /root/.ssh/ && touch /root/.ssh/known_hosts && ssh-keyscan -t rsa bitbucket.org >> /root/.ssh/known_hosts
RUN --mount=type=secret,id=ssh,target=/root/.ssh/id_rsa go mod download

COPY . .

RUN go build

# Runtime Container

FROM alpine:3.14

RUN apk add --no-cache libc6-compat tzdata

COPY --from=builder /go/src/latonaio/azure-face-api-registrator-golang-kube .

CMD ["./azure-face-api-registrator-golang-kube"]
