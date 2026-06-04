.PHONY: build clean deps

# Build all services
build:
	go build -o tracking-api.exe ./service/tracking/api
	go build -o tracking-srv.exe ./service/tracking/rpc

# Clean binaries
clean:
	rm -f tracking-api.exe tracking-srv.exe

# Update dependencies
deps:
	cd service/tracking/api && go mod tidy
	cd service/tracking/rpc && go mod tidy
	go work sync