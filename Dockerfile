FROM golang:1.21 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -o sfs
RUN chmod +x ./sfs
RUN ./sfs --help

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/sfs .
EXPOSE 8080

CMD ["sfs"]
