.PHONY: up build watch scale k8s-apply

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