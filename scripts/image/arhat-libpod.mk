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
image.build.arhat-libpod.linux.x86:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.amd64:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.armv6:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.armv7:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.arm64:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.ppc64le:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.s390x:
	sh scripts/image/build.sh $@

image.build.arhat-libpod.linux.all: \
	image.build.arhat-libpod.linux.amd64 \
	image.build.arhat-libpod.linux.arm64 \
	image.build.arhat-libpod.linux.armv7 \
	image.build.arhat-libpod.linux.armv6 \
	image.build.arhat-libpod.linux.s390x \
	image.build.arhat-libpod.linux.ppc64le
	# image.build.arhat-libpod.linux.x86 \

# push
image.push.arhat-libpod.linux.x86:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.amd64:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.armv6:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.armv7:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.arm64:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.ppc64le:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.s390x:
	sh scripts/image/push.sh $@

image.push.arhat-libpod.linux.all: \
	image.push.arhat-libpod.linux.amd64 \
	image.push.arhat-libpod.linux.arm64 \
	image.push.arhat-libpod.linux.armv7 \
	image.push.arhat-libpod.linux.armv6 \
	image.push.arhat-libpod.linux.s390x \
	image.push.arhat-libpod.linux.ppc64le
	# image.push.arhat-libpod.linux.x86 \
