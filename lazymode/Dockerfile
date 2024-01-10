# Use an official Golang runtime as a parent image
FROM golang:alpine

# Update the package list and install bash
RUN apk update && apk upgrade && apk add --no-cache bash

# Set the working directory inside the container
WORKDIR /app

# Copy the local code to the container's working directory
COPY . /app

# Build the Go application
RUN go build -o main .

# Check if the application binary exists and is executable
RUN ls -alh

# Expose port 4567 for the application
EXPOSE 4200

# Copy the startup script into the container and give execution rights
COPY startup.sh startup.sh
RUN chmod +x startup.sh

# Use the startup script as the entry point
CMD ["./startup.sh"]

