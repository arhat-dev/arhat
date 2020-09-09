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
template-application-go:
	sh scripts/build/build.sh $@

# linux
template-application-go.linux.x86:
	sh scripts/build/build.sh $@

template-application-go.linux.amd64:
	sh scripts/build/build.sh $@

template-application-go.linux.armv5:
	sh scripts/build/build.sh $@

template-application-go.linux.armv6:
	sh scripts/build/build.sh $@

template-application-go.linux.armv7:
	sh scripts/build/build.sh $@

template-application-go.linux.arm64:
	sh scripts/build/build.sh $@

template-application-go.linux.mips:
	sh scripts/build/build.sh $@

template-application-go.linux.mipshf:
	sh scripts/build/build.sh $@

template-application-go.linux.mipsle:
	sh scripts/build/build.sh $@

template-application-go.linux.mipslehf:
	sh scripts/build/build.sh $@

template-application-go.linux.mips64:
	sh scripts/build/build.sh $@

template-application-go.linux.mips64hf:
	sh scripts/build/build.sh $@

template-application-go.linux.mips64le:
	sh scripts/build/build.sh $@

template-application-go.linux.mips64lehf:
	sh scripts/build/build.sh $@

template-application-go.linux.ppc64:
	sh scripts/build/build.sh $@

template-application-go.linux.ppc64le:
	sh scripts/build/build.sh $@

template-application-go.linux.s390x:
	sh scripts/build/build.sh $@

template-application-go.linux.riscv64:
	sh scripts/build/build.sh $@

template-application-go.linux.all: \
	template-application-go.linux.x86 \
	template-application-go.linux.amd64 \
	template-application-go.linux.armv5 \
	template-application-go.linux.armv6 \
	template-application-go.linux.armv7 \
	template-application-go.linux.arm64 \
	template-application-go.linux.mips \
	template-application-go.linux.mipshf \
	template-application-go.linux.mipsle \
	template-application-go.linux.mipslehf \
	template-application-go.linux.mips64 \
	template-application-go.linux.mips64hf \
	template-application-go.linux.mips64le \
	template-application-go.linux.mips64lehf \
	template-application-go.linux.ppc64 \
	template-application-go.linux.ppc64le \
	template-application-go.linux.s390x \
	template-application-go.linux.riscv64

template-application-go.darwin.amd64:
	sh scripts/build/build.sh $@

# template-application-go.darwin.arm64:
# 	sh scripts/build/build.sh $@

template-application-go.darwin.all: \
	template-application-go.darwin.amd64

template-application-go.windows.x86:
	sh scripts/build/build.sh $@

template-application-go.windows.amd64:
	sh scripts/build/build.sh $@

# template-application-go.windows.arm64:
# 	sh scripts/build/build.sh $@

template-application-go.windows.all: \
	template-application-go.windows.x86 \
	template-application-go.windows.amd64
