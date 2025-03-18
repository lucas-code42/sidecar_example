APP_IMAGE=base64-http:latest
SIDECAR_IMAGE=sidecar:latest

up:
	minikube start
	minikube image load $(SIDECAR_IMAGE)
	minikube image load $(APP_IMAGE)
	kubectl apply -f pod.yaml

build:
	docker build -t $(APP_IMAGE) .
	docker build -t $(SIDECAR_IMAGE) ./sidecar/.

restart:
	kubectl delete pod base64-pod --ignore-not-found
	kubectl apply -f pod.yaml

logs:
	kubectl logs base64-pod -c base64-http
	kubectl logs base64-pod -c sidecar

status:
	kubectl get pods


port-forward:
	kubectl port-forward base64-pod 8080:8080

clean:
	kubectl delete pod base64-pod --ignore-not-found
	minikube delete
	docker rmi $(APP_IMAGE) $(SIDECAR_IMAGE) --force
