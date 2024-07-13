build:
	@go build -o bin/dc
run: build
	@./bin/dc
test:
	@go test ./... -v