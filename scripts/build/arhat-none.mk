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
	arhat-none.linux.s390x

arhat-none.darwin.amd64:
	sh scripts/build/build.sh $@

# # currently darwin/arm64 build will fail due to golang link error
# arhat-none.darwin.arm64:
# 	sh scripts/build/build.sh $@

arhat-none.darwin.all: \
	arhat-none.darwin.amd64

arhat-none.windows.x86:
	sh scripts/build/build.sh $@

arhat-none.windows.amd64:
	sh scripts/build/build.sh $@

# currently windows/arm builds do not support metrics collecting due to
# https://github.com/go-ole/go-ole/issues/202
arhat-none.windows.armv5:
	sh scripts/build/build.sh $@

arhat-none.windows.armv6:
	sh scripts/build/build.sh $@

arhat-none.windows.armv7:
	sh scripts/build/build.sh $@

# # currently no support for windows/arm64
# arhat-none.windows.arm64:
# 	sh scripts/build/build.sh $@

arhat-none.windows.all: \
	arhat-none.windows.x86 \
	arhat-none.windows.amd64

# arhat-none.android.amd64:
# 	sh scripts/build/build.sh $@

# arhat-none.android.x86:
# 	sh scripts/build/build.sh $@

# arhat-none.android.armv5:
# 	sh scripts/build/build.sh $@

# arhat-none.android.armv6:
# 	sh scripts/build/build.sh $@

# arhat-none.android.armv7:
# 	sh scripts/build/build.sh $@

# arhat-none.android.arm64:
# 	sh scripts/build/build.sh $@

# arhat-none.android.all: \
# 	arhat-none.android.amd64 \
# 	arhat-none.android.arm64 \
# 	arhat-none.android.x86 \
# 	arhat-none.android.armv7 \
# 	arhat-none.android.armv5 \
# 	arhat-none.android.armv6

arhat-none.freebsd.amd64:
	sh scripts/build/build.sh $@

arhat-none.freebsd.x86:
	sh scripts/build/build.sh $@

arhat-none.freebsd.armv5:
	sh scripts/build/build.sh $@

arhat-none.freebsd.armv6:
	sh scripts/build/build.sh $@

arhat-none.freebsd.armv7:
	sh scripts/build/build.sh $@

arhat-none.freebsd.arm64:
	sh scripts/build/build.sh $@

arhat-none.freebsd.all: \
	arhat-none.freebsd.amd64 \
	arhat-none.freebsd.arm64 \
	arhat-none.freebsd.armv7 \
	arhat-none.freebsd.x86 \
	arhat-none.freebsd.armv5 \
	arhat-none.freebsd.armv6

arhat-none.netbsd.amd64:
	sh scripts/build/build.sh $@

arhat-none.netbsd.x86:
	sh scripts/build/build.sh $@

arhat-none.netbsd.armv5:
	sh scripts/build/build.sh $@

arhat-none.netbsd.armv6:
	sh scripts/build/build.sh $@

arhat-none.netbsd.armv7:
	sh scripts/build/build.sh $@

arhat-none.netbsd.arm64:
	sh scripts/build/build.sh $@

arhat-none.netbsd.all: \
	arhat-none.netbsd.amd64 \
	arhat-none.netbsd.arm64 \
	arhat-none.netbsd.armv7 \
	arhat-none.netbsd.x86 \
	arhat-none.netbsd.armv5 \
	arhat-none.netbsd.armv6

arhat-none.openbsd.amd64:
	sh scripts/build/build.sh $@

arhat-none.openbsd.x86:
	sh scripts/build/build.sh $@

arhat-none.openbsd.armv5:
	sh scripts/build/build.sh $@

arhat-none.openbsd.armv6:
	sh scripts/build/build.sh $@

arhat-none.openbsd.armv7:
	sh scripts/build/build.sh $@

arhat-none.openbsd.arm64:
	sh scripts/build/build.sh $@

arhat-none.openbsd.all: \
	arhat-none.openbsd.amd64 \
	arhat-none.openbsd.arm64 \
	arhat-none.openbsd.armv7 \
	arhat-none.openbsd.x86 \
	arhat-none.openbsd.armv5 \
	arhat-none.openbsd.armv6

arhat-none.solaris.amd64:
	sh scripts/build/build.sh $@

arhat-none.aix.ppc64:
	sh scripts/build/build.sh $@

arhat-none.dragonfly.amd64:
	sh scripts/build/build.sh $@

arhat-none.plan9.amd64:
	sh scripts/build/build.sh $@

arhat-none.plan9.x86:
	sh scripts/build/build.sh $@

arhat-none.plan9.armv5:
	sh scripts/build/build.sh $@

arhat-none.plan9.armv6:
	sh scripts/build/build.sh $@

arhat-none.plan9.armv7:
	sh scripts/build/build.sh $@

arhat-none.plan9.all: \
	arhat-none.plan9.amd64 \
	arhat-none.plan9.armv7 \
	arhat-none.plan9.x86 \
	arhat-none.plan9.armv5 \
	arhat-none.plan9.armv6
