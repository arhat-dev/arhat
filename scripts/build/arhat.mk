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
arhat:
	sh scripts/build/build.sh $@

# any platform
arhat.%.%:
	sh scripts/build/build.sh $@

# linux
arhat.linux.x86:
	sh scripts/build/build.sh $@

arhat.linux.amd64:
	sh scripts/build/build.sh $@

arhat.linux.armv5:
	sh scripts/build/build.sh $@

arhat.linux.armv6:
	sh scripts/build/build.sh $@

arhat.linux.armv7:
	sh scripts/build/build.sh $@

arhat.linux.arm64:
	sh scripts/build/build.sh $@

arhat.linux.mips:
	sh scripts/build/build.sh $@

arhat.linux.mipshf:
	sh scripts/build/build.sh $@

arhat.linux.mipsle:
	sh scripts/build/build.sh $@

arhat.linux.mipslehf:
	sh scripts/build/build.sh $@

arhat.linux.mips64:
	sh scripts/build/build.sh $@

arhat.linux.mips64hf:
	sh scripts/build/build.sh $@

arhat.linux.mips64le:
	sh scripts/build/build.sh $@

arhat.linux.mips64lehf:
	sh scripts/build/build.sh $@

arhat.linux.ppc64:
	sh scripts/build/build.sh $@

arhat.linux.ppc64le:
	sh scripts/build/build.sh $@

arhat.linux.s390x:
	sh scripts/build/build.sh $@

arhat.linux.riscv64:
	sh scripts/build/build.sh $@

arhat.linux.all: \
	arhat.linux.x86 \
	arhat.linux.amd64 \
	arhat.linux.armv5 \
	arhat.linux.armv6 \
	arhat.linux.armv7 \
	arhat.linux.arm64 \
	arhat.linux.riscv64 \
	arhat.linux.mips \
	arhat.linux.mipshf \
	arhat.linux.mipsle \
	arhat.linux.mipslehf \
	arhat.linux.mips64 \
	arhat.linux.mips64hf \
	arhat.linux.mips64le \
	arhat.linux.mips64lehf \
	arhat.linux.ppc64 \
	arhat.linux.ppc64le \
	arhat.linux.s390x

arhat.darwin.amd64:
	sh scripts/build/build.sh $@

# # currently darwin/arm64 build will fail due to golang link error
# arhat.darwin.arm64:
# 	sh scripts/build/build.sh $@

arhat.darwin.all: \
	arhat.darwin.amd64

arhat.windows.x86:
	sh scripts/build/build.sh $@

arhat.windows.amd64:
	sh scripts/build/build.sh $@

arhat.windows.armv5:
	sh scripts/build/build.sh $@

arhat.windows.armv6:
	sh scripts/build/build.sh $@

arhat.windows.armv7:
	sh scripts/build/build.sh $@

# # currently no support for windows/arm64
# arhat.windows.arm64:
# 	sh scripts/build/build.sh $@

arhat.windows.all: \
	arhat.windows.x86 \
	arhat.windows.amd64 \
	arhat.windows.armv5 \
	arhat.windows.armv6 \
	arhat.windows.armv7

# arhat.android.amd64:
# 	sh scripts/build/build.sh $@

# arhat.android.x86:
# 	sh scripts/build/build.sh $@

# arhat.android.armv5:
# 	sh scripts/build/build.sh $@

# arhat.android.armv6:
# 	sh scripts/build/build.sh $@

# arhat.android.armv7:
# 	sh scripts/build/build.sh $@

# arhat.android.arm64:
# 	sh scripts/build/build.sh $@

# arhat.android.all: \
# 	arhat.android.amd64 \
# 	arhat.android.arm64 \
# 	arhat.android.x86 \
# 	arhat.android.armv7 \
# 	arhat.android.armv5 \
# 	arhat.android.armv6

arhat.freebsd.amd64:
	sh scripts/build/build.sh $@

arhat.freebsd.x86:
	sh scripts/build/build.sh $@

arhat.freebsd.armv5:
	sh scripts/build/build.sh $@

arhat.freebsd.armv6:
	sh scripts/build/build.sh $@

arhat.freebsd.armv7:
	sh scripts/build/build.sh $@

arhat.freebsd.arm64:
	sh scripts/build/build.sh $@

arhat.freebsd.all: \
	arhat.freebsd.amd64 \
	arhat.freebsd.arm64 \
	arhat.freebsd.armv7 \
	arhat.freebsd.x86 \
	arhat.freebsd.armv5 \
	arhat.freebsd.armv6

arhat.netbsd.amd64:
	sh scripts/build/build.sh $@

arhat.netbsd.x86:
	sh scripts/build/build.sh $@

arhat.netbsd.armv5:
	sh scripts/build/build.sh $@

arhat.netbsd.armv6:
	sh scripts/build/build.sh $@

arhat.netbsd.armv7:
	sh scripts/build/build.sh $@

arhat.netbsd.arm64:
	sh scripts/build/build.sh $@

arhat.netbsd.all: \
	arhat.netbsd.amd64 \
	arhat.netbsd.arm64 \
	arhat.netbsd.armv7 \
	arhat.netbsd.x86 \
	arhat.netbsd.armv5 \
	arhat.netbsd.armv6

arhat.openbsd.amd64:
	sh scripts/build/build.sh $@

arhat.openbsd.x86:
	sh scripts/build/build.sh $@

arhat.openbsd.armv5:
	sh scripts/build/build.sh $@

arhat.openbsd.armv6:
	sh scripts/build/build.sh $@

arhat.openbsd.armv7:
	sh scripts/build/build.sh $@

arhat.openbsd.arm64:
	sh scripts/build/build.sh $@

arhat.openbsd.all: \
	arhat.openbsd.amd64 \
	arhat.openbsd.arm64 \
	arhat.openbsd.armv7 \
	arhat.openbsd.x86 \
	arhat.openbsd.armv5 \
	arhat.openbsd.armv6

arhat.solaris.amd64:
	sh scripts/build/build.sh $@

arhat.aix.ppc64:
	sh scripts/build/build.sh $@

arhat.dragonfly.amd64:
	sh scripts/build/build.sh $@

arhat.plan9.amd64:
	sh scripts/build/build.sh $@

arhat.plan9.x86:
	sh scripts/build/build.sh $@

arhat.plan9.armv5:
	sh scripts/build/build.sh $@

arhat.plan9.armv6:
	sh scripts/build/build.sh $@

arhat.plan9.armv7:
	sh scripts/build/build.sh $@

arhat.plan9.all: \
	arhat.plan9.amd64 \
	arhat.plan9.armv7 \
	arhat.plan9.x86 \
	arhat.plan9.armv5 \
	arhat.plan9.armv6

# TODO: currently tinygo lacks json and tls support
#
# docker run \
# 	-v $(shell pwd):/app/src/arhat.dev/arhat \
# 	-v $(shell pwd)/vendor:/go/src \
# 	-e "GOPATH=/go:/app" \
# 	tinygo/tinygo:0.15.0 \
# 	tinygo build -o \
# 		/app/src/arhat.dev/arhat/build/arhat.js.wasm \
# 		-target wasm -tags 'nometrics' \
# 		--no-debug \
# 		/go/src/arhat.dev/arhat/cmd/arhat
arhat.js.wasm:
	sh scripts/build/build.sh $@
