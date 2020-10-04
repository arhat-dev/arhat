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

# build
image.build.arhat-docker.linux.x86:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.amd64:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.armv5:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.armv6:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.armv7:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.arm64:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.ppc64le:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.mips64le:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.s390x:
	sh scripts/image/build.sh $@

image.build.arhat-docker.linux.all: \
	image.build.arhat-docker.linux.amd64 \
	image.build.arhat-docker.linux.arm64 \
	image.build.arhat-docker.linux.armv7 \
	image.build.arhat-docker.linux.armv6 \
	image.build.arhat-docker.linux.armv5 \
	image.build.arhat-docker.linux.x86 \
	image.build.arhat-docker.linux.s390x \
	image.build.arhat-docker.linux.ppc64le \
	image.build.arhat-docker.linux.mips64le

image.build.arhat-docker.windows.amd64:
	sh scripts/image/build.sh $@

image.build.arhat-docker.windows.armv7:
	sh scripts/image/build.sh $@

image.build.arhat-docker.windows.all: \
	image.build.arhat-docker.windows.amd64 \
	image.build.arhat-docker.windows.armv7

# push
image.push.arhat-docker.linux.x86:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.amd64:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.armv5:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.armv6:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.armv7:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.arm64:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.ppc64le:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.mips64le:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.s390x:
	sh scripts/image/push.sh $@

image.push.arhat-docker.linux.all: \
	image.push.arhat-docker.linux.amd64 \
	image.push.arhat-docker.linux.arm64 \
	image.push.arhat-docker.linux.armv7 \
	image.push.arhat-docker.linux.armv6 \
	image.push.arhat-docker.linux.armv5 \
	image.push.arhat-docker.linux.x86 \
	image.push.arhat-docker.linux.s390x \
	image.push.arhat-docker.linux.ppc64le \
	image.push.arhat-docker.linux.mips64le

image.push.arhat-docker.windows.amd64:
	sh scripts/image/push.sh $@

image.push.arhat-docker.windows.armv7:
	sh scripts/image/push.sh $@

image.push.arhat-docker.windows.all: \
	image.push.arhat-docker.windows.amd64 \
	image.push.arhat-docker.windows.armv7
