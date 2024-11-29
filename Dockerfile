# Use the official Go image as the base image
FROM golang:1.20-alpine

WORKDIR /app

# Copy the Go source code and go.mod files into the container
COPY . .

# Install dependencies (for Go modules)
RUN go mod tidy

# Build the Go application
RUN go build -o bitcoin-collateral .

# Expose port 8080 for the app
EXPOSE 8080

# Set the command to run the application
CMD ["./bitcoin-collateral"]
