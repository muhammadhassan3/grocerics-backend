
FROM golang:1.26-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd


FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
USER app
COPY --from=build /out/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["api"]
