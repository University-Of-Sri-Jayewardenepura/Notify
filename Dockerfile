FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /notify ./cmd/notify

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates wget
RUN addgroup -S notify && adduser -S notify -G notify

COPY --from=build /notify /usr/local/bin/notify

EXPOSE 8080

USER notify

CMD ["notify"]
