REGISTRY      ?= tmaxcloudck
VERSION       ?= 0.0.1

HS_IMG   = $(REGISTRY)/template-validating-webhook:$(VERSION)

.PHONY: test build push

# Test apis func
test:
	go test ./pkg/apis

# Build the docker image
build:
	docker build -f build/Dockerfile -t $(HS_IMG) . 

# Push the docker image 
push:
	docker push $(HS_IMG)