#!/bin/bash
NGINX_CONF=/etc/nginx/conf.d/default.conf

# Set backend endpoint, defaulting to "http://backend:4567" if not specified
if [ -z "$YELB_APPSERVER_ENDPOINT" ]; then YELB_APPSERVER_ENDPOINT="http://backend:4567"; fi

# Add search domain to resolv.conf if SEARCH_DOMAIN is specified
if [ $SEARCH_DOMAIN ]; then echo "search ${SEARCH_DOMAIN}" >> /etc/resolv.conf; fi

# Create a new Nginx configuration
{
    echo 'server {'
    echo '    listen       80;'
    echo '    server_name  localhost;'
    echo '    root /usr/share/nginx/html;'  # Explicitly set root for static files

    # Set client_max_body_size
    echo '    client_max_body_size 10M;'

    # Configure reverse proxy for /api
    echo '    location /api {'
    echo "        proxy_pass $YELB_APPSERVER_ENDPOINT;"
    echo '        proxy_http_version 1.1;'
    echo '        proxy_set_header Upgrade $http_upgrade;'
    echo '        proxy_set_header Connection "upgrade";'
    echo '        gzip on;'
    echo '        gzip_types text/plain text/css application/json application/javascript application/x-javascript text/xml application/xml application/xml+rss text/javascript;'
    echo '        gunzip on;'
    echo '    }'

    echo '}'
} > $NGINX_CONF

# Start Nginx in non-daemon mode
nginx -g "daemon off;"
