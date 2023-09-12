.PHONY: build test clean prepare update docker

GO = CGO_ENABLED=1 go

MICROSERVICES=cmd/sample-service

.PHONY: $(MICROSERVICES)

DOCKERS=docker_sample_service_go
.PHONY: $(DOCKERS)

VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)
GIT_SHA=$(shell git rev-parse HEAD)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/sample-service.Version=$(VERSION)"

GOTESTFLAGS ?=-race

tidy:
	go mod tidy

build: $(MICROSERVICES)

cmd/sample-service:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test: unit_test
	$(GO) vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]
	./bin/test-attribution-txt.sh

unit_test:
ifneq (,$(findstring json,$(GOTESTFLAGS)))
	$(GO) install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	$(GO) test $(GOTESTFLAGS) ./... -coverprofile=coverage.out | gotestfmt
else
	$(GO) test $(GOTESTFLAGS) ./... -coverprofile=coverage.out
endif

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

vendor:
	go mod vendor