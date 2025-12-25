FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o writer .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/writer /app/writer
EXPOSE 50051
ENTRYPOINT ["/app/writer"]
