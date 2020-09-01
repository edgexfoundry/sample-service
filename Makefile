.PHONY: build test clean prepare update docker

GO = CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/sample-service

.PHONY: $(MICROSERVICES)

DOCKERS=docker_sample_service_go
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
GIT_SHA=$(shell git rev-parse HEAD)
GOFLAGS=-ldflags "-X github.com/edgexfoundry/sample-service.Version=$(VERSION)"

build: $(MICROSERVICES)
	$(GO) build ./...

cmd/sample-service:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test:
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./bin/test-attribution-txt.sh
	./bin/test-go-mod-tidy.sh

clean:
	rm -f $(MICROSERVICES)

docker: $(DOCKERS)

docker_sample_service_go:
	docker build \
		--build-arg http_proxy=$(http_proxy) \
		--build-arg https_proxy=$(https_proxy) \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-sample-service-go:$(GIT_SHA) \
		-t edgexfoundry/docker-sample-service-go:$(VERSION)-dev \
		.
