FROM golang:tip-alpine3.24 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=0 go build -o main main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
EXPOSE 8080
ENTRYPOINT ["./main"]
