FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN apt-get update && apt-get install -y gcc
ENV CGO_ENABLED=1 GOOS=linux
RUN go build -ldflags="-w -s" -o server .


FROM debian:stable-slim
WORKDIR /root/
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /app/server .

CMD ["./server"]