# Copyright 2020 The arhat.dev Authors.
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

# IMAGE_REPOS is the comma separated list of image registries
IMAGE_REPOS ?= docker.io/arhatdev,ghcr.io/arhat-dev

export IMAGE_REPOS

DEFAULT_IMAGE_MANIFEST_TAG ?= latest

include scripts/lint.mk

GOMOD := GOPROXY=direct GOSUMDB=off go mod
.PHONY: vendor
vendor:
	${GOMOD} tidy
	${GOMOD} vendor

# testing
include scripts/test/unit.mk

# binary build
include scripts/build/arhat-none.mk
include scripts/build/arhat-docker.mk
include scripts/build/arhat-libpod.mk

# image
include scripts/image/arhat-none.mk
include scripts/image/arhat-docker.mk
include scripts/image/arhat-libpod.mk

image.build.linux.all: \
	image.build.arhat-none.linux.all \
	image.build.arhat-docker.linux.all \
	image.build.arhat-libpod.linux.all

image.build.windows.all: \
	image.build.arhat-none.windows.all \
	image.build.arhat-docker.windows.all

image.push.linux.all: \
	image.push.arhat-none.linux.all \
	image.push.arhat-docker.linux.all \
	image.push.arhat-libpod.linux.all

image.push.windows.all: \
	image.push.arhat-none.windows.all \
	image.push.arhat-docker.windows.all

# packaging
include scripts/package/arhat-none.mk
include scripts/package/arhat-docker.mk
include scripts/package/arhat-libpod.mk

# optional private scripts
-include private/scripts.mk
