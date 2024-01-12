#use the official Python image as the base image
FROM python:3.9

#set the working directory
WORKDIR /app

#copy the application files into the container
COPY . /app

#create the uploads directory
RUN mkdir -p /app/uploads


#install required packages
#RUN pip install --no-cache-dir flask==2.0.1
#RUN pip install cloudone-vsapi
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install visionone-filesecurity

#set the environment variables (only availabe in us-1 region right now will leave as default) please just add your api key
#ENV C1_ADDRESS="antimalware.us-1.cloudone.trendmicro.com:443"
#ENV C1_API_KEY="YOUR C1 API KEY HERE"
#ENV C1_REGION="us-1"

#expose port 5000
EXPOSE 5000

#run the Flask application
CMD ["python", "app.py"]