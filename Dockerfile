FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o runebird ./cmd/emailer

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/runebird .

COPY emailer.yaml .

RUN mkdir -p templates logs

EXPOSE 8080

CMD ["./runebird"]