TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=rafay
NAME=rafay
BINARY=terraform-provider-${NAME}
VERSION=1.1.7
GIT_BRANCH ?= main
OS_ARCH=darwin_amd64
BUCKET_NAME ?= terraform-provider-rafay
BUILD_NUMBER ?= $(shell date "+%Y%m%d-%H%M")
TAG := $(or $(shell git describe --tags --exact-match  2>/dev/null), $(shell echo "origin/${GIT_BRANCH}"))

default: install

build:
	export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ${BINARY}
	#go generate

release:
	GOOS=darwin GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

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
