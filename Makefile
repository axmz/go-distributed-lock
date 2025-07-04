.PHONY: build watch ka kd kc kl install-protoc

build:
	docker compose build

watch:
	docker compose up --scale app=5 --watch

dev: build watch

ka:
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/redis.yml
	kubectl wait --for=condition=ready pod -l app=redis --timeout=60s
	kubectl apply -f k8s/go-verifier.yml
	kubectl wait --for=condition=ready pod -l app=verifier --timeout=60s
	kubectl apply -f k8s/go-distrilock.yml

kd:
	kubectl delete -f k8s/

kc:
	kubectl create configmap distrilock-env --from-env-file=.env

kl:
	kubectl get pods -o name | xargs -n1 kubectl logs

install-protoc:
	sudo apt update
	sudo apt install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$$PATH:$(go env GOPATH)/bin"