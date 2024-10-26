# Use the official Golang image
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Install wait-for-it
ADD https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh /usr/local/bin/wait-for-it
RUN chmod +x /usr/local/bin/wait-for-it

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go get github.com/lib/pq

# Copy the source code
COPY . .

# Expose port $APPROVAL_WEBSERVER_PORT
EXPOSE $APPROVAL_WEBSERVER_PORT

# Update the entrypoint to wait for postgres
ENTRYPOINT ["wait-for-it", "postgres:5432", "--"]
