.PHONY: up build watch scale k8s-apply install-protoc

build:
	docker compose build

up:
	docker compose up -d

watch:
	docker compose watch

scale:
	docker compose up -d --scale app=5

dev: build up scale watch

k8s-apply:
	kubectl apply -f k8s/

k8s-delete:
	kubectl delete -f k8s/

k8s-configmap:
	kubectl create configmap distrilock-env --from-env-file=.env

install-protoc:
	sudo apt update
	sudo apt install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$$PATH:$(go env GOPATH)/bin"