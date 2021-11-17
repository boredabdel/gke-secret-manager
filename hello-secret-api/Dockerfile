## Build
FROM golang:1.16-alpine as build

RUN apk add --no-cache git

WORKDIR /tmp/app

# Download necessary Go modules
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/hello-secret-api .
##
## Deploy
##
FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build /tmp/app/out/hello-secret-api /app/hello-secret-api

EXPOSE 8080

ENTRYPOINT ["/app/hello-secret-api"]