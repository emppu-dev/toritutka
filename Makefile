build:
	@go build -o bin/golang-api main.go

run: build
	@./bin/golang-api