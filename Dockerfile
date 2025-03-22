# Use the official Go image as a parent image
FROM golang:1.23.2-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN go build -o main cmd/api/main.go

# Use a smaller image for the final stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .
# Copy any necessary configuration files
COPY --from=builder /app/config* ./

COPY .env ./
# Expose the port your application uses
EXPOSE 8080

# Command to run the application
CMD ["./main"]