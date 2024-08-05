FROM golang:1.20 as builder

RUN mkdir /app

ADD . /app/

WORKDIR /app

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add sqlite

WORKDIR /app

COPY --from=builder /app/main .

COPY config.yaml /app/config.yaml

COPY db.sqlite /app/db.sqlite

CMD ["/app/main"]
