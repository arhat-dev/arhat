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

#
# linux
#
package.arhat-libpod.deb.amd64:
	sh scripts/package/package.sh $@

package.arhat-libpod.deb.armv6:
	sh scripts/package/package.sh $@

package.arhat-libpod.deb.armv7:
	sh scripts/package/package.sh $@

package.arhat-libpod.deb.arm64:
	sh scripts/package/package.sh $@

package.arhat-libpod.deb.all: \
	package.arhat-libpod.deb.amd64 \
	package.arhat-libpod.deb.armv6 \
	package.arhat-libpod.deb.armv7 \
	package.arhat-libpod.deb.arm64

package.arhat-libpod.rpm.amd64:
	sh scripts/package/package.sh $@

package.arhat-libpod.rpm.armv7:
	sh scripts/package/package.sh $@

package.arhat-libpod.rpm.arm64:
	sh scripts/package/package.sh $@

package.arhat-libpod.rpm.all: \
	package.arhat-libpod.rpm.amd64 \
	package.arhat-libpod.rpm.armv7 \
	package.arhat-libpod.rpm.arm64

package.arhat-libpod.linux.all: \
	package.arhat-libpod.deb.all \
	package.arhat-libpod.rpm.all
