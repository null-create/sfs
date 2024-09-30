# Build
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify && go mod tidy
COPY . .
RUN GOOS=linux go build -o sfs
RUN chmod +x ./sfs 
RUN ./sfs --help
RUN ./sfs client --new && ./sfs server --new

# Prepare image
FROM ubuntu:22.04
RUN apt-get update -y && apt-get install -y \
  ca-certificates \
  sqlite3 \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app .
RUN echo "contents: " && ls -la
EXPOSE 8080

CMD ["./sfs", "server", "--start"]
