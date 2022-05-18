REGISTRY      ?= tmaxcloudck
VERSION       ?= 0.0.1

HS_IMG   = $(REGISTRY)/helm-apiserver:$(VERSION)

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


# Custom target for Helm API server
.PHONY: test-lint test-unit

# Test code lint
test-lint:
# golangci-lint run ./... -v -E gofmt --timeout 1h0m0s
	golint ./...

# Unit test
test-unit:
	go test -v ./pkg/apis/...
