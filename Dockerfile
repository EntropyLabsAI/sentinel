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

# Expose port $APPROVAL_WEBSERVER_PORT
EXPOSE $APPROVAL_WEBSERVER_PORT
