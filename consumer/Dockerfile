FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /consumer ./main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /consumer .

CMD ["/root/consumer"]
