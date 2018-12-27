PACKAGE      = devdashboard
VERSION      = $(shell git log -n1 --pretty='%h')
BUILD_DIR    = build
RELEASE_DIR  = dist
RELEASE_FILE = $(PACKAGE)_$(VERSION)_$(shell go env GOOS)-$(shell go env GOARCH)
REPO         = github.com/urld/devdashboard

.PHONY: all clean clean_build clean_dist dist build install test generate


all: test install dist



dist: build
	mkdir -p $(RELEASE_DIR)
	mkdir -p $(BUILD_DIR)/licenses
	cp LICENSE $(BUILD_DIR)/licenses/devdashboard.LICENSE
	cp -r cmd/devdashboard/static $(BUILD_DIR)/static
	cp -r cmd/devdashboard/templates $(BUILD_DIR)/templates
	tar -czf  $(RELEASE_DIR)/$(RELEASE_FILE).tar.gz $(BUILD_DIR) --transform='s/$(BUILD_DIR)/$(RELEASE_FILE)/g'


build: clean_build generate
	mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && \
	go build -v $(REPO)/cmd/devdashboard


generate:
	go generate -v $(REPO)/devdashpb


test: generate
	go test -race -v $(REPO)/...

coverage: generate
	go test -race -v -cover -coverprofile=coverage.out $(REPO)/...
	go tool cover -html=coverage.out -o coverage.html

fixture: generate
	go run cmd/devdashfixture/main.go

install: generate
	go install -v $(REPO)/cmd/devdashboard


clean: clean_build clean_dist


clean_build:
	rm -rf $(BUILD_DIR)


clean_dist:
	rm -rf $(RELEASE_DIR)

