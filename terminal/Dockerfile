# Use an official Ubuntu runtime as a parent image
FROM golang:latest

# Set noninteractive installation mode
ENV DEBIAN_FRONTEND=noninteractive

# Set the working directory inside the container
WORKDIR /app

# Copy the local code to the container's working directory
COPY . /app

# Build the Go application
RUN go build -o main .

# Expose port 8081 for the application
EXPOSE 8081

# Update the package list and install necessary packages
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    openssh-client \
    sudo && \
    rm -rf /var/lib/apt/lists/*

# Create AWS credentials file
RUN mkdir -p /root/.aws && \
    echo "[default]\naws_access_key_id=AKIAIOSFODNN7PROD\naws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYPRODKEY" > /root/.aws/credentials && \
    chmod 600 /root/.aws/credentials

# Create .ssh directory and test key file, and set appropriate permissions
RUN mkdir -p /root/.ssh && \
    touch /root/.ssh/test_key_file && \
    chmod 700 /root/.ssh && \
    chmod 600 /root/.ssh/test_key_file

# Inserting and encoding the message directly into test_key_file
RUN echo "Congratulations on navigating through the maze of the Boring Paper Company!\nHere's your reward for all the hard work: FLAG{P4per_Trail_Discovered!}\nRemember, sometimes the most unassuming surfaces hold the deepest secrets." | base64 > /root/.ssh/test_key_file

# Set the working directory inside the container
WORKDIR /app

# Copy the local code to the container's working directory
COPY . /app

# Build the Go application
RUN go build -o main .

# Expose port 8081 for the application
EXPOSE 8081

# Define the command to run the app
CMD ["./main"]

