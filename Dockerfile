FROM golang:1.21 as builder

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY . .
RUN go build -o sfs

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/sfs .
EXPOSE 8080

CMD ["./sfs"]
