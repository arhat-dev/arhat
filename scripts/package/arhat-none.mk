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
package.arhat-none.deb.amd64:
	sh scripts/package/package.sh $@

package.arhat-none.deb.armv6:
	sh scripts/package/package.sh $@

package.arhat-none.deb.armv7:
	sh scripts/package/package.sh $@

package.arhat-none.deb.arm64:
	sh scripts/package/package.sh $@

package.arhat-none.deb.all: \
	package.arhat-none.deb.amd64 \
	package.arhat-none.deb.armv6 \
	package.arhat-none.deb.armv7 \
	package.arhat-none.deb.arm64

package.arhat-none.rpm.amd64:
	sh scripts/package/package.sh $@

package.arhat-none.rpm.armv7:
	sh scripts/package/package.sh $@

package.arhat-none.rpm.arm64:
	sh scripts/package/package.sh $@

package.arhat-none.rpm.all: \
	package.arhat-none.rpm.amd64 \
	package.arhat-none.rpm.armv7 \
	package.arhat-none.rpm.arm64

package.arhat-none.linux.all: \
	package.arhat-none.deb.all \
	package.arhat-none.rpm.all

#
# windows
#

package.arhat-none.msi.amd64:
	sh scripts/package/package.sh $@

package.arhat-none.msi.arm64:
	sh scripts/package/package.sh $@

package.arhat-none.msi.all: \
	package.arhat-none.msi.amd64 \
	package.arhat-none.msi.arm64

package.arhat-none.windows.all: \
	package.arhat-none.msi.all

#
# darwin
#

package.arhat-none.pkg.amd64:
	sh scripts/package/package.sh $@

package.arhat-none.pkg.arm64:
	sh scripts/package/package.sh $@

package.arhat-none.pkg.all: \
	package.arhat-none.pkg.amd64 \
	package.arhat-none.pkg.arm64

package.arhat-none.darwin.all: \
	package.arhat-none.pkg.all
