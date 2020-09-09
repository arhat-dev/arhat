# Copyright Go-IIoT (https://github.com/goiiot)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Enable oneshell
#.ONESHELL:

# Use custom interpreter
# SHELL = /bin/bash
# .SHELLFLAGS = -c

-include ../common/os.mk

# app binary output name
BINARY = libserial
# add ext name (.exe) when running in windows
ifeq ($(NATIVE_OS),windows)
	BINARY := $(BINARY).exe
endif

# app main source file
FILE_BINARY_SOURCE = cmd/libserial/main.go
# args for go test -run or go test -bench
RUN ?= .
# app arguments
ARGS ?= 

DIR_BUILD = .build
DIR_DIST = .dist
DIR_TEST = .test

$(DIR_TEST):
	mkdir -p $(DIR_TEST)

$(DIR_BUILD):
	mkdir -p $(DIR_BUILD)

$(DIR_DIST):
	mkdir -p $(DIR_DIST)

#
# Tools
#

GO = go
GOFMT = gofmt
GIT = git
UPX = $(shell which upx)
DOCKER = $(shell which docker)

ifneq ($(filter not found,$(GO)),)
    GO = :
else ifeq ($(GO),)
    GO = :
endif

ifneq ($(filter not found,$(GOFMT)),)
    GOFMT = :
else ifeq ($(GOFMT),)
    GOFMT = :
endif

ifneq ($(filter not found,$(GIT)),)
    GIT = :
else ifeq ($(GIT),)
    GIT = :
endif

ifneq ($(filter not found,$(UPX)),)
    UPX = :
else ifeq ($(UPX),)
    UPX = :
endif

ifneq ($(filter not found,$(DOCKER)),)
    DOCKER = :
else ifeq ($(DOCKER),)
    DOCKER = :
endif

#
# Version and git info
#

GOVERSION = $(shell $(GO) version | cut -d\  -f3)
COMMIT = $(shell $(GIT) rev-parse HEAD 2>/dev/null | tail -c 16)
BRANCH = $(shell $(GIT) rev-parse --abbrev-ref HEAD 2>/dev/null)
VERSION = $(shell $(GIT) describe --tags 2>/dev/null)

ifeq ($(COMMIT),HEAD)
	COMMIT = none
else ifeq ($(COMMIT),)
	COMMIT = none
endif

ifeq ($(BRANCH),HEAD)
	BRANCH = none
else ifeq ($(BRANCH),)
	BRANCH = none
endif

ifeq ($(VERSION),)
	VERSION = none
endif

#
# Build and run
#

FILE_BINARY = $(DIR_BUILD)/$(BINARY)
FILE_RELEASE_BINARY = $(DIR_DIST)/$(BINARY)

LDFLAGS := \
	-X main.branch=$(BRANCH) \
	-X main.commit=$(COMMIT) \
	-X main.version=$(VERSION) \
	-X main.goVersion=$(GOVERSION)

BUILD_ARGS := -ldflags '$(LDFLAGS)'

RELEASE_LDFLAGS := -w -s $(LDFLAGS)

RELEASE_FLAGS := -ldflags='$(RELEASE_LDFLAGS)'

.PHONY: build release run

build: $(DIR_BUILD)
	$(GO) build $(BUILD_ARGS) -o $(FILE_BINARY) $(FILE_BINARY_SOURCE)

run: build
	./$(FILE_BINARY) $(ARGS)

release: $(DIR_DIST)
	$(GO) build -buildmode=pie $(RELEASE_FLAGS) -o $(FILE_RELEASE_BINARY) $(FILE_BINARY_SOURCE)
	$(UPX) --brute $(FILE_RELEASE_BINARY)

#
# Test and coverage
#

FILE_TEST_BIN = $(DIR_TEST)/$(BINARY).test
FILE_COVERAGE = coverage.txt

TEST_FLAGS := \
	-v -race \
	-covermode=atomic \
	-coverprofile=$(FILE_COVERAGE)

.PHONY: test coverage

pty_start:
	socat -d -d pty,raw,echo=0 pty,raw,echo=0 > socat.out 2>&1 & echo $$! > socat.pid

pty_stop:
	kill -KILL `cat socat.pid 2>/dev/null` 2>/dev/null || true
	rm -f socat.pid socat.out

test: $(DIR_TEST)
	$(GO) test $(TEST_FLAGS) -run=$(RUN) $(PKG) -args $(ARGS)

coverage: test
	$(GO) tool cover -html=$(FILE_COVERAGE)

#
# Benchmark and profiling
#

FILE_BLOCK_PROFILE = $(DIR_TEST)/blockprofile.out
FILE_CPU_PROFILE = $(DIR_TEST)/cpuprofile.out
FILE_MEM_PROFILE = $(DIR_TEST)/memprofile.out
FILE_MUTEX_PROFILE = $(DIR_TEST)/mutexprofile.out
FILE_TRACE_PROFILE = $(DIR_TEST)/trace.out

BENCH_FLAGS := \
	-v -race -benchmem \
	-trace=$(FILE_TRACE_PROFILE) \
	-blockprofile=$(FILE_BLOCK_PROFILE) \
	-cpuprofile=$(FILE_CPU_PROFILE) \
	-memprofile=$(FILE_MEM_PROFILE) \
	-mutexprofile=$(FILE_MUTEX_PROFILE) \
	-o=$(FILE_TEST_BIN)

.PHONY: benchmark profile_cpu profile_mem profile_block profile_trace \
		profile_all_start profile_all_stop

benchmark: $(DIR_TEST)
	$(GO) test $(BENCH_FLAGS) -run=Benchmark.* -bench=$(RUN) $(PKG)

profile_cpu:
	$(GO) tool pprof -http localhost:50080 $(FILE_TEST_BIN) $(FILE_CPU_PROFILE)

profile_mem:
	$(GO) tool pprof -http localhost:50081 $(FILE_TEST_BIN) $(FILE_MEM_PROFILE)

profile_block:
	$(GO) tool pprof -http localhost:50082 $(FILE_TEST_BIN) $(FILE_BLOCK_PROFILE)

profile_mutex:
	$(GO) tool pprof -http localhost:50083 $(FILE_TEST_BIN) $(FILE_MUTEX_PROFILE)

profile_trace:
	$(GO) tool trace -http localhost:50084 $(FILE_TEST_BIN) $(FILE_TRACE_PROFILE)

FILE_CPU_PROFILE_PID = $(DIR_TEST)/cpuprofile.pid
FILE_MEM_PROFILE_PID = $(DIR_TEST)/memprofile.pid
FILE_BLOCK_PROFILE_PID = $(DIR_TEST)/blockprofile.pid
FILE_MUTEX_PROFILE_PID = $(DIR_TEST)/mutexprofile.pid
FILE_TRACE_PROFILE_PID = $(DIR_TEST)/trace.pid

profile_all_start:
	$(GO) tool pprof -http localhost:50080 $(FILE_TEST_BIN) $(FILE_CPU_PROFILE) \
		& echo $$! > $(FILE_CPU_PROFILE_PID)
	$(GO) tool pprof -http localhost:50081 $(FILE_TEST_BIN) $(FILE_MEM_PROFILE) \
		& echo $$! > $(FILE_MEM_PROFILE_PID)
	$(GO) tool pprof -http localhost:50082 $(FILE_TEST_BIN) $(FILE_BLOCK_PROFILE) \
		& echo $$! > $(FILE_BLOCK_PROFILE_PID)
	$(GO) tool pprof -http localhost:50083 $(FILE_TEST_BIN) $(FILE_MUTEX_PROFILE) \
		& echo $$! > $(FILE_MUTEX_PROFILE_PID)
	$(GO) tool trace -http localhost:50084 $(FILE_TEST_BIN) $(FILE_TRACE_PROFILE) \
		& echo $$! > $(FILE_TRACE_PROFILE_PID)

profile_all_stop: \
	$(FILE_BLOCK_PROFILE_PID) $(FILE_CPU_PROFILE_PID) \
	$(FILE_MEM_PROFILE_PID) $(FILE_MUTEX_PROFILE_PID) \
	$(FILE_TRACE_PROFILE_PID)

	kill -KILL `cat $(FILE_BLOCK_PROFILE_PID)`
	kill -KILL `cat $(FILE_CPU_PROFILE_PID)`
	kill -KILL `cat $(FILE_MEM_PROFILE_PID)`
	kill -KILL `cat $(FILE_MUTEX_PROFILE_PID)`
	kill -KILL `cat $(FILE_TRACE_PROFILE_PID)`

	rm -f $(FILE_BLOCK_PROFILE_PID)
	rm -f $(FILE_CPU_PROFILE_PID)
	rm -f $(FILE_MEM_PROFILE_PID)
	rm -f $(FILE_MUTEX_PROFILE_PID)
	rm -f $(FILE_TRACE_PROFILE_PID)

#
# Format
#

.PHONY: fmt style_check

fmt:
	gofmt -s -w .

style_check: $(DIR_TEST)
	gofmt -s -e -d . | tee $(DIR_TEST)/fmt.out
	grep -q 'diff' $(DIR_TEST)/fmt.out; test $$? -eq 1

#
# Dependency management
#

.PHONY: tidy ensure

tidy:
	$(GO) mod tidy

ensure:
	$(GO) mod download

#
# Image build
#
TAG ?= $(BINARY):latest
CONTEXT ?= .
DOCKERFILE ?= Dockerfile

.PHONY: image run_image

image:
	$(DOCKER) build -t $(TAG) -f $(DOCKERFILE) $(CONTEXT)

run_image:
	$(DOCKER) run --rm -it $(TAG)

#
# Cleanup
#

.PHONY: clean

clean:
	rm -rf $(DIR_DIST)
	rm -rf $(DIR_BUILD)
	rm -rf $(DIR_TEST)
