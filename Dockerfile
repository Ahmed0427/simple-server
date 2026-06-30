FROM golang:tip-alpine3.24 AS build
WORKDIR /app
COPY main.go .
RUN go build -o main main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
EXPOSE 8080
ENTRYPOINT ["./main"]
