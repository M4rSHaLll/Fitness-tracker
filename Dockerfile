FROM golang:1.25.6-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/fitness-tracker ./cmd/telegram

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

COPY --from=build /out/fitness-tracker /usr/local/bin/fitness-tracker

USER app
EXPOSE 8080

ENTRYPOINT ["fitness-tracker"]
