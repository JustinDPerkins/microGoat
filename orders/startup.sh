#!/bin/bash

# Define the default port
DEFAULT_PORT=5000

# Set the port environment variable if not already set
export PORT=${PORT:-$DEFAULT_PORT}

# when the variable is populated, a search domain entry is added to resolv.conf at startup
# this is needed for the ECS service discovery as the app works by calling host names and not FQDNs
# a search domain can't be added to the container when using the awsvpc mode 
# and the awsvpc mode is needed for A records (bridge only supports SRV records)
if [ -n "${SEARCH_DOMAIN}" ]; then
    echo "search ${SEARCH_DOMAIN}" >> /etc/resolv.conf
fi

# Start the Go application on the specified port
./main -port=${PORT}