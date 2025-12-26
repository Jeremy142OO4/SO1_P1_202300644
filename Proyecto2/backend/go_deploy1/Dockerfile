FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o go-deploy1 .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/go-deploy1 /app/go-deploy1
EXPOSE 8080
ENTRYPOINT ["/app/go-deploy1"]