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
package.arhat.deb.amd64:
	sh scripts/package/package.sh $@

package.arhat.deb.armv6:
	sh scripts/package/package.sh $@

package.arhat.deb.armv7:
	sh scripts/package/package.sh $@

package.arhat.deb.arm64:
	sh scripts/package/package.sh $@

package.arhat.deb.all: \
	package.arhat.deb.amd64 \
	package.arhat.deb.armv6 \
	package.arhat.deb.armv7 \
	package.arhat.deb.arm64

package.arhat.rpm.amd64:
	sh scripts/package/package.sh $@

package.arhat.rpm.armv7:
	sh scripts/package/package.sh $@

package.arhat.rpm.arm64:
	sh scripts/package/package.sh $@

package.arhat.rpm.all: \
	package.arhat.rpm.amd64 \
	package.arhat.rpm.armv7 \
	package.arhat.rpm.arm64

package.arhat.linux.all: \
	package.arhat.deb.all \
	package.arhat.rpm.all

#
# windows
#

package.arhat.msi.amd64:
	sh scripts/package/package.sh $@

package.arhat.msi.arm64:
	sh scripts/package/package.sh $@

package.arhat.msi.all: \
	package.arhat.msi.amd64 \
	package.arhat.msi.arm64

package.arhat.windows.all: \
	package.arhat.msi.all

#
# darwin
#

package.arhat.pkg.amd64:
	sh scripts/package/package.sh $@

package.arhat.pkg.arm64:
	sh scripts/package/package.sh $@

package.arhat.pkg.all: \
	package.arhat.pkg.amd64 \
	package.arhat.pkg.arm64

package.arhat.darwin.all: \
	package.arhat.pkg.all
