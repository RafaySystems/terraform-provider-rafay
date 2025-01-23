TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=rafay
NAME=rafay
BINARY=terraform-provider-${NAME}
VERSION=1.1.28
GIT_BRANCH ?= main
OS_ARCH := $(shell uname | grep -q 'Linux' && echo "linux_amd64" || echo "darwin_amd64")
BUCKET_NAME ?= terraform-provider-rafay
BUILD_NUMBER ?= $(shell date "+%Y%m%d-%H%M")
TAG := $(or $(shell git describe --tags --exact-match  2>/dev/null), $(shell echo "origin/${GIT_BRANCH}"))

default: install

build:
	export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ${BINARY}
	#go generate

release:
	GOOS=darwin GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_darwin_amd64 -buildvcs=false
	GOOS=darwin GOARCH=arm64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_darwin_arm64 -buildvcs=false
	GOOS=freebsd GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_freebsd_386 -buildvcs=false
	GOOS=freebsd GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_freebsd_amd64 -buildvcs=false
	GOOS=freebsd GOARCH=arm GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_freebsd_arm -buildvcs=false
	GOOS=linux GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_386 -buildvcs=false
	GOOS=linux GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_amd64 -buildvcs=false
	GOOS=linux GOARCH=arm GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_arm -buildvcs=false
	GOOS=linux GOARCH=arm64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_linux_arm64 -buildvcs=false
	GOOS=openbsd GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_openbsd_386 -buildvcs=false
	GOOS=openbsd GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_openbsd_amd64 -buildvcs=false
	GOOS=solaris GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_solaris_amd64 -buildvcs=false
	GOOS=windows GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_windows_386 -buildvcs=false
	GOOS=windows GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore CGO_ENABLED=0 go build -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./bin/${BINARY}_${VERSION}_windows_amd64 -buildvcs=false

zip:
	$(shell cd bin;	zip ${BINARY}_${VERSION}_linux_arm64.zip ${BINARY}_${VERSION}_linux_arm64)
	$(shell cd bin; zip ${BINARY}_${VERSION}_linux_amd64.zip ${BINARY}_${VERSION}_linux_amd64)
	$(shell cd bin; zip ${BINARY}_${VERSION}_darwin_arm64.zip ${BINARY}_${VERSION}_darwin_arm64)
	$(shell cd bin; zip ${BINARY}_${VERSION}_darwin_amd64.zip ${BINARY}_${VERSION}_darwin_amd64)

install: build
	bash internal/scripts/fwgen.sh
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m


fwgen:
	bash internal/scripts/fwgen.sh
	
push:
	aws s3 cp ./bin/${BINARY}_${VERSION}_darwin_amd64  s3://$(BUCKET_NAME)/$(TAG)/$(BUILD_NUMBER)/${BINARY}_${VERSION}_darwin_amd64 --no-progress
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
	GOPRIVATE=github.com/RafaySystems/* go get -d github.com/RafaySystems/rafay-common@master

.PHONY: test-migrate
test-migrate:
	go test -v ./rafay/migrate/...
