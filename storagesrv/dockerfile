FROM golang:1.16.10-bullseye
 
WORKDIR /app
COPY . .
 
RUN apt-get update && \
  apt-get install git && \
  go get github.com/cespare/reflex
 
EXPOSE 7000-8000
CMD ["reflex", "-c", "reflex.conf"]
