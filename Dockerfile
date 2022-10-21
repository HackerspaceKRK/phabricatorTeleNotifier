FROM golang:alpine AS builder


# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git bash && mkdir -p /build/phabricatorTeleNotifier

WORKDIR /build/phabricatorTeleNotifier

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download -json

COPY . .

RUN mkdir -p /app && CGO_ENABLED=0 GOOS=${TARGETPLATFORM%%/*} GOARCH=${TARGETPLATFORM##*/} \
    go build -ldflags='-s -w -extldflags="-static"' -o /app/phabricatorTeleNotifier

FROM scratch AS bin-unix
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/phabricatorTeleNotifier /app/phabricatorTeleNotifier

LABEL org.opencontainers.image.description A docker image for the phabricatorTeleNotifier telegram bot.

ENTRYPOINT ["/app/phabricatorTeleNotifier"]
