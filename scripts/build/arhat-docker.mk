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

# native
arhat-docker:
	sh scripts/build/build.sh $@

# linux
arhat-docker.linux.x86:
	sh scripts/build/build.sh $@

arhat-docker.linux.amd64:
	sh scripts/build/build.sh $@

arhat-docker.linux.armv5:
	sh scripts/build/build.sh $@

arhat-docker.linux.armv6:
	sh scripts/build/build.sh $@

arhat-docker.linux.armv7:
	sh scripts/build/build.sh $@

arhat-docker.linux.arm64:
	sh scripts/build/build.sh $@

arhat-docker.linux.mips:
	sh scripts/build/build.sh $@

arhat-docker.linux.mipshf:
	sh scripts/build/build.sh $@

arhat-docker.linux.mipsle:
	sh scripts/build/build.sh $@

arhat-docker.linux.mipslehf:
	sh scripts/build/build.sh $@

arhat-docker.linux.mips64:
	sh scripts/build/build.sh $@

arhat-docker.linux.mips64hf:
	sh scripts/build/build.sh $@

arhat-docker.linux.mips64le:
	sh scripts/build/build.sh $@

arhat-docker.linux.mips64lehf:
	sh scripts/build/build.sh $@

arhat-docker.linux.ppc64:
	sh scripts/build/build.sh $@

arhat-docker.linux.ppc64le:
	sh scripts/build/build.sh $@

arhat-docker.linux.s390x:
	sh scripts/build/build.sh $@

arhat-docker.linux.riscv64:
	sh scripts/build/build.sh $@

arhat-docker.linux.all: \
	arhat-docker.linux.x86 \
	arhat-docker.linux.amd64 \
	arhat-docker.linux.armv5 \
	arhat-docker.linux.armv6 \
	arhat-docker.linux.armv7 \
	arhat-docker.linux.arm64 \
	arhat-docker.linux.mips \
	arhat-docker.linux.mipshf \
	arhat-docker.linux.mipsle \
	arhat-docker.linux.mipslehf \
	arhat-docker.linux.mips64 \
	arhat-docker.linux.mips64hf \
	arhat-docker.linux.mips64le \
	arhat-docker.linux.mips64lehf \
	arhat-docker.linux.ppc64 \
	arhat-docker.linux.ppc64le \
	arhat-docker.linux.s390x \
	arhat-docker.linux.riscv64

arhat-docker.darwin.amd64:
	sh scripts/build/build.sh $@

# arhat-docker.darwin.arm64:
# 	sh scripts/build/build.sh $@

arhat-docker.darwin.all: \
	arhat-docker.darwin.amd64

arhat-docker.windows.x86:
	sh scripts/build/build.sh $@

arhat-docker.windows.amd64:
	sh scripts/build/build.sh $@

arhat-docker.windows.armv6:
	sh scripts/build/build.sh $@

arhat-docker.windows.armv7:
	sh scripts/build/build.sh $@

# arhat-docker.windows.arm64:
# 	sh scripts/build/build.sh $@

arhat-docker.windows.all: \
	arhat-docker.windows.x86 \
	arhat-docker.windows.amd64
