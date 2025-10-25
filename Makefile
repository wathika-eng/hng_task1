@run: go run .

# build and remove debug info
@build: go build -ldflags="-s -w" -o hng_task1 .

@tidy: go mod tidy
