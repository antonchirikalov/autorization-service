FROM golang:1.12-alpine
WORKDIR /usr/src/app
COPY ./dist/ .
EXPOSE 8080
CMD ./api
