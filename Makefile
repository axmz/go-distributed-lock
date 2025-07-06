.PHONY: build watch run install-protoc clean

# Docker Compose commands
build:
	docker compose build

watch:
	docker compose up --scale app=5 --watch

run: clean build watch

# Development setup
install-protoc:
	sudo apt update
	sudo apt install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$$PATH:$(go env GOPATH)/bin"

# Cleanup
clean:
	docker compose down
	docker compose down --volumes --remove-orphans