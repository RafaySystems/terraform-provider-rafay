TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
TOFU_HOSTNAME=registry.opentofu.org
NAMESPACE=rafaysystems
NAME=rafay
BINARY=terraform-provider-${NAME}
VERSION=1.1.28
GIT_BRANCH ?= main
OS := $(shell uname | grep -q 'Linux' && echo "linux" || echo "darwin")
ARCH := $(shell uname -m | grep -q 'x86_64' && echo "amd64" || echo "arm64")
OS_ARCH := ${OS}_${ARCH}
BUCKET_NAME ?= terraform-provider-rafay
BUILD_NUMBER ?= $(shell date "+%Y%m%d-%H%M")
TAG := $(or $(shell git describe --tags --exact-match  2>/dev/null), $(shell echo "origin/${GIT_BRANCH}"))

default: install

build:
	export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ${BINARY}
	#go generate

release:
	GOOS=darwin GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_darwin_amd64 -buildvcs=false -p 2
	GOOS=darwin GOARCH=arm64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_darwin_arm64 -buildvcs=false -p 2
	GOOS=freebsd GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_freebsd_386 -buildvcs=false -p 2
	GOOS=freebsd GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_freebsd_amd64 -buildvcs=false -p 2
	GOOS=freebsd GOARCH=arm GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_freebsd_arm -buildvcs=false -p 2
	GOOS=linux GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_386 -buildvcs=false -p 2
	GOOS=linux GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_amd64 -buildvcs=false -p 2
	GOOS=linux GOARCH=arm GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_arm -buildvcs=false -p 2
	GOOS=linux GOARCH=arm64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_arm64 -buildvcs=false -p 2
	GOOS=openbsd GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_openbsd_386 -buildvcs=false -p 2
	GOOS=openbsd GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_openbsd_amd64 -buildvcs=false -p 2
	GOOS=solaris GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_solaris_amd64 -buildvcs=false -p 2
	GOOS=windows GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_windows_386 -buildvcs=false -p 2
	GOOS=windows GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_windows_amd64 -buildvcs=false -p 2

zip:
	$(shell cd bin;	zip ${BINARY}_${VERSION}_linux_arm64.zip ${BINARY}_${VERSION}_linux_arm64)
	$(shell cd bin; zip ${BINARY}_${VERSION}_linux_amd64.zip ${BINARY}_${VERSION}_linux_amd64)
	$(shell cd bin; zip ${BINARY}_${VERSION}_darwin_arm64.zip ${BINARY}_${VERSION}_darwin_arm64)
	$(shell cd bin; zip ${BINARY}_${VERSION}_darwin_amd64.zip ${BINARY}_${VERSION}_darwin_amd64)

install: build
	bash internal/scripts/fwgen.sh
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mkdir -p ~/.terraform.d/plugins/${TOFU_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${TOFU_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall:
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	rm -rf ~/.terraform.d/plugins/${TOFU_HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

# New organized test targets based on reorganized test structure
.PHONY: test-unit test-integration test-plan-only test-negative test-framework test-all-organized

test-unit:
	@echo "Running unit tests (internal function tests in rafay/ package)..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay

test-integration:
	@echo "Running all integration tests..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=integration ./tests/integration/...

test-plan-only:
	@echo "Running plan-only integration tests..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=planonly ./tests/integration/plan_only/

test-negative:
	@echo "Running negative integration tests..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags='!planonly' ./tests/integration/negative/

test-framework:
	@echo "Running Plugin Framework tests..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=planonly ./tests/framework/

test-all-organized:
	@echo "Running all tests with organized structure..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay ./tests/...

# Test targets with coverage
.PHONY: test-unit-cover test-integration-cover test-all-cover

test-unit-cover:
	@echo "Running unit tests with coverage..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover ./rafay

test-integration-cover:
	@echo "Running integration tests with coverage..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover -tags=integration ./tests/integration/...

test-all-cover:
	@echo "Running all tests with coverage..."
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover ./rafay ./tests/...


fwgen:
	bash internal/scripts/fwgen.sh

push:
	aws s3 cp ./bin/${BINARY}_${VERSION}_darwin_amd64  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_darwin_amd64 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_darwin_arm64  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_darwin_arm64 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_freebsd_386  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_freebsd_386 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_freebsd_amd64  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_freebsd_amd64 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_freebsd_arm  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_freebsd_arm --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_linux_386  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_linux_386 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_linux_amd64  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_linux_amd64 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_linux_arm  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_linux_arm --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_openbsd_386  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_openbsd_386 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_solaris_amd64  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_solaris_amd64 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_windows_386 s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_windows_386 --no-progress
	aws s3 cp ./bin/${BINARY}_${VERSION}_windows_amd64 s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_windows_amd64 --no-progress

bucket-name:
	echo 'Build Folder URL:- https://$(BUCKET_NAME).s3.us-west-1.amazonaws.com/$(TAG)/$(BUILD_NUMBER)/'


.PHONY: tidy
tidy:
	GOPRIVATE=github.com/RafaySystems/* go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: update-deps
update-deps:
	GOPRIVATE=github.com/RafaySystems/* go get github.com/RafaySystems/rafay-common@master

.PHONY: test-migrate
test-migrate:
	go test -v ./rafay/migrate/...

.PHONY: generate
generate:
	cd tools; go generate ./...

.PHONY: clean
clean:
	rm -r ./bin
