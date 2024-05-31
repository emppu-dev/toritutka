build:
	@go build -o bin/toritutka main.go

run: build
	@./bin/toritutka