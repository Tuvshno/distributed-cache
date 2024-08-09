build:
	@go build -o bin/dc
run: build
	@./bin/dc

runfollower: build
	@./bin/dc --listenaddr :4000 --leaderaddr :3000

test:
	@go test -count=1 ./... -v || { echo 'Tests failed'; exit 1; }