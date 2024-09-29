# Use the official Golang image
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["go", "run", "main.go", "handlers.go", "websockets.go"]
