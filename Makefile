.PHONY: run install cleanup-helm cleanup-namespace cleanup logs watch build-distrilock build-verifier build push-distrilock push-verifier push

run: cleanup-helm cleanup-namespace install logs

watch:
	@echo "Following verifier logs in real-time..."
	@echo "Press Ctrl+C to stop following logs"
	@kubectl logs -f -l app=verifier -n distrilock-redis-cluster

cleanup-helm:
	@echo "Uninstalling Helm release..."
	@helm uninstall distrilock --namespace distrilock-redis-cluster --ignore-not-found=true || true

cleanup-namespace:
	@echo "Deleting namespace..."
	@kubectl delete namespace distrilock-redis-cluster --ignore-not-found=true || true

install:
	@echo "Installing distrilock in distrilock-redis-cluster namespace..."
	@helm install distrilock ./chart --namespace distrilock-redis-cluster --create-namespace --wait

logs:
	@echo "Following verifier logs in real-time..."
	@echo "Press Ctrl+C to stop following logs"
	@echo "Waiting for verifier to be ready..."
	@while ! kubectl logs -f -l app=verifier -n distrilock-redis-cluster 2>/dev/null; do \
		echo "Verifier not ready yet, retrying in 1 second..."; \
		sleep 1; \
	done

build: build-distrilock build-verifier

build-distrilock:
	docker build -f Dockerfile.distrilock -t axmz/distrilock:cluster .

build-verifier:
	docker build -f Dockerfile.verifier -t axmz/verifier:cluster .

push: push-distrilock push-verifier

push-distrilock:
	docker push axmz/distrilock:cluster

push-verifier:
	docker push axmz/verifier:cluster