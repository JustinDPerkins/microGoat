# Use an official Golang runtime as a parent image
FROM golang:alpine

# Update the package list and install bash and curl
RUN apk update && apk upgrade && apk add --no-cache bash curl

# Set the working directory inside the container
WORKDIR /app

# Copy the local code to the container's working directory
COPY . /app

# Build the Go application
RUN go build -o main .

# Check if the application binary exists and is executable
RUN ls -alh

# Expose port 4567 for the application
EXPOSE 4567

# Copy the startup script into the container and give execution rights
COPY startup.sh startup.sh
RUN chmod +x startup.sh

# Create .ssh directory and test key file, and set appropriate permissions
RUN mkdir -p /root/.ssh && touch /root/.ssh/test_key_file && chmod 700 /root/.ssh && chmod 600 /root/.ssh/test_key_file

# Create AWS credentials file
RUN mkdir -p /root/.aws && \
    echo -e "[default]\naws_access_key_id=AKIAIOSFODNN7EXAMPLE\naws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" > /root/.aws/credentials && \
    chmod 600 /root/.aws/credentials

# Use the startup script as the entry point
CMD ["./startup.sh"]

