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
arhat-libpod:
	sh scripts/build/build.sh $@

# linux
arhat-libpod.linux.x86:
	sh scripts/build/build.sh $@

arhat-libpod.linux.amd64:
	sh scripts/build/build.sh $@

arhat-libpod.linux.armv5:
	sh scripts/build/build.sh $@

arhat-libpod.linux.armv6:
	sh scripts/build/build.sh $@

arhat-libpod.linux.armv7:
	sh scripts/build/build.sh $@

arhat-libpod.linux.arm64:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mips:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mipshf:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mipsle:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mipslehf:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mips64:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mips64hf:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mips64le:
	sh scripts/build/build.sh $@

arhat-libpod.linux.mips64lehf:
	sh scripts/build/build.sh $@

arhat-libpod.linux.ppc64:
	sh scripts/build/build.sh $@

arhat-libpod.linux.ppc64le:
	sh scripts/build/build.sh $@

arhat-libpod.linux.s390x:
	sh scripts/build/build.sh $@

arhat-libpod.linux.riscv64:
	sh scripts/build/build.sh $@

arhat-libpod.linux.all: \
	arhat-libpod.linux.x86 \
	arhat-libpod.linux.amd64 \
	arhat-libpod.linux.armv5 \
	arhat-libpod.linux.armv6 \
	arhat-libpod.linux.armv7 \
	arhat-libpod.linux.arm64 \
	arhat-libpod.linux.mips \
	arhat-libpod.linux.mipshf \
	arhat-libpod.linux.mipsle \
	arhat-libpod.linux.mipslehf \
	arhat-libpod.linux.mips64 \
	arhat-libpod.linux.mips64hf \
	arhat-libpod.linux.mips64le \
	arhat-libpod.linux.mips64lehf \
	arhat-libpod.linux.ppc64 \
	arhat-libpod.linux.ppc64le \
	arhat-libpod.linux.s390x \
	arhat-libpod.linux.riscv64
