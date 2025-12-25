FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o consumidor .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/consumidor /app/consumidor
ENTRYPOINT ["/app/consumidor"]
