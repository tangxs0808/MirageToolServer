FROM golang:1.20 as builder

RUN mkdir /app

ADD . /app/

WORKDIR /app

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

COPY config /app/config

CMD ["/app/main"]
