FROM golang:1.22.5-alpine3.20
WORKDIR /app
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN go build -o bin .
ENTRYPOINT ["/app/bin"]