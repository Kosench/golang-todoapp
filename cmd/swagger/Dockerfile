FROM golang:1.25.9-alpine3.23
RUN apk add --no-cache git
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
WORKDIR /code
ENTRYPOINT ["swag"]