
default: all

all:  check-fmt install_deps vet test


.PHONY: vendor
vendor:
	GOPROXY=direct GOPRIVATE=github.com/RafaySystems/* go mod vendor

generate:
	go generate ./...
	$(MAKE) check

.PHONY: tidy
tidy:
	GOPROXY=direct GOPRIVATE=github.com/RafaySystems/* go mod tidy

.PHONY: check
check:
	go fmt ./...
	go vet ./...

	$(MAKE) tidy

IMG ?= rctl:latest-${BUILD_NUMBER}


.PHONY: generate_info_file
generate_info_file:
	echo "TAG=$(TAG)" > $(BUILD_INFO_FILE)
	echo "ORG=$(ORG)" >> $(BUILD_INFO_FILE)
	echo "TIME=$(TS)" >> $(BUILD_INFO_FILE)

.PHONY: docker-build
docker-build:
	docker build . -t ${IMG} \
		--build-arg BUILD_USR=${BUILD_USER} \
		--build-arg BUILD_PWD=${BUILD_PASSWORD} \
		--build-arg VERSION=${VERSION} \
		--build-arg DATE_TIME="${DATE_TIME}" \
		--build-arg BUILD_NUM="${BUILD_NUMBER}" \
		--build-arg GIT_BRANCH="${GIT_BRANCH}" \
		;

.PHONY: build
build: docker-build
	$(MAKE) generate_info_file
	./build-tools/copy_from_docker.sh ${IMG} ${BUILD_FOLDER}

test: install_deps vet
	$(info ******************** running tests ********************)
	go test -v -coverprofile=coverage.out ./...

.PHONY: install_deps
install_deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...

.PHONY: vet
vet:
	$(info ******************** vetting ********************)
	go vet ./...

.PHONY: check-fmt
check-fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

.PHONY: fmt
fmt:
	$(info ******************** running formatting ********************)
	@test -z $(shell gofmt -w $(SRC)) || exit 1

.PHONY: clean
clean:
	$(info ******************** cleaning compiled binaries ********************)
	rm -rf $(BIN)

.PHONY: install
install: check-fmt vet
	$(info ******************** installing binary ********************)
	go build -ldflags=$(LDFLAGS) -o bin/rctl
	$(info packing binary)
	upx bin/rctl
	mv bin/rctl ${GOPATH}/bin

# not used by the Jenkins build. Used to upload binaries to github release
.PHONY: compile
compile:
	$(info ******************** compiling binaries ********************)
	gox \
		-output=$(BIN)/{{.Dir}}_{{.OS}}_{{.Arch}} \
		-ldflags=$(LDFLAGS) \
		;


