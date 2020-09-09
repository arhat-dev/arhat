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
arhat-none:
	sh scripts/build/build.sh $@

# linux
arhat-none.linux.x86:
	sh scripts/build/build.sh $@

arhat-none.linux.amd64:
	sh scripts/build/build.sh $@

arhat-none.linux.armv5:
	sh scripts/build/build.sh $@

arhat-none.linux.armv6:
	sh scripts/build/build.sh $@

arhat-none.linux.armv7:
	sh scripts/build/build.sh $@

arhat-none.linux.arm64:
	sh scripts/build/build.sh $@

arhat-none.linux.mips:
	sh scripts/build/build.sh $@

arhat-none.linux.mipshf:
	sh scripts/build/build.sh $@

arhat-none.linux.mipsle:
	sh scripts/build/build.sh $@

arhat-none.linux.mipslehf:
	sh scripts/build/build.sh $@

arhat-none.linux.mips64:
	sh scripts/build/build.sh $@

arhat-none.linux.mips64hf:
	sh scripts/build/build.sh $@

arhat-none.linux.mips64le:
	sh scripts/build/build.sh $@

arhat-none.linux.mips64lehf:
	sh scripts/build/build.sh $@

arhat-none.linux.ppc64:
	sh scripts/build/build.sh $@

arhat-none.linux.ppc64le:
	sh scripts/build/build.sh $@

arhat-none.linux.s390x:
	sh scripts/build/build.sh $@

arhat-none.linux.riscv64:
	sh scripts/build/build.sh $@

arhat-none.linux.all: \
	arhat-none.linux.x86 \
	arhat-none.linux.amd64 \
	arhat-none.linux.armv5 \
	arhat-none.linux.armv6 \
	arhat-none.linux.armv7 \
	arhat-none.linux.arm64 \
	arhat-none.linux.mips \
	arhat-none.linux.mipshf \
	arhat-none.linux.mipsle \
	arhat-none.linux.mipslehf \
	arhat-none.linux.mips64 \
	arhat-none.linux.mips64hf \
	arhat-none.linux.mips64le \
	arhat-none.linux.mips64lehf \
	arhat-none.linux.ppc64 \
	arhat-none.linux.ppc64le \
	arhat-none.linux.s390x \
	arhat-none.linux.riscv64

arhat-none.darwin.amd64:
	sh scripts/build/build.sh $@

# arhat-none.darwin.arm64:
# 	sh scripts/build/build.sh $@

arhat-none.darwin.all: \
	arhat-none.darwin.amd64

arhat-none.windows.x86:
	sh scripts/build/build.sh $@

arhat-none.windows.amd64:
	sh scripts/build/build.sh $@

arhat-none.windows.armv6:
	sh scripts/build/build.sh $@

arhat-none.windows.armv7:
	sh scripts/build/build.sh $@

# arhat-none.windows.arm64:
# 	sh scripts/build/build.sh $@

arhat-none.windows.all: \
	arhat-none.windows.x86 \
	arhat-none.windows.amd64
