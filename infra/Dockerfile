FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/bot ./cmd/bot

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=build /bin/bot /bin/bot

ENTRYPOINT ["/bin/bot"]
