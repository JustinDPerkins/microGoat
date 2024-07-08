# Use an official Golang runtime as a parent image
FROM golang:latest

# Set noninteractive installation mode
ENV DEBIAN_FRONTEND=noninteractive

# Set the working directory inside the container
WORKDIR /app

# Copy the local code to the container's working directory
COPY . /app
COPY entrypoint.sh /entrypoint.sh

# Update the package list and install necessary packages including cron and openssl
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    openssh-client \
    sudo \
    cron \
    openssl && \
    rm -rf /var/lib/apt/lists/*

# Build the Go application
RUN go build -o main .

# Make the entrypoint script executable
RUN chmod +x /entrypoint.sh

# Expose port 8081 for the application
EXPOSE 8081

# Set the entrypoint to the entrypoint script
ENTRYPOINT ["/entrypoint.sh"]

