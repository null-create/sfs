# Start from the official Golang image as the build environment
FROM golang:1.21 as builder

# Set the working directory inside the container
WORKDIR /app
COPY go.mod go.sum ./

# Download the Go modules and build
RUN go mod download
COPY . .
RUN go build -o sfs

# Start a new stage from a minimal base image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/sfs .
EXPOSE 8080

# Command to run the application
CMD ["./sfs"]
