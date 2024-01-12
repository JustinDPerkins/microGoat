FROM nginx:alpine

# Update the package list and install bash
RUN apk update && apk upgrade && apk add --no-cache bash

COPY project/ /usr/share/nginx/html/

EXPOSE 80

COPY project/startup.sh startup.sh
RUN chmod +x startup.sh

CMD ["./startup.sh"]