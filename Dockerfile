FROM golang:alpine3.17 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/main .

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/main .
COPY font font
CMD ["./main"]
