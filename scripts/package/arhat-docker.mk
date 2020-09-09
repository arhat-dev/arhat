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
package.arhat-docker.deb.amd64:
	sh scripts/package/package.sh $@

package.arhat-docker.deb.armv6:
	sh scripts/package/package.sh $@

package.arhat-docker.deb.armv7:
	sh scripts/package/package.sh $@

package.arhat-docker.deb.arm64:
	sh scripts/package/package.sh $@

package.arhat-docker.deb.all: \
	package.arhat-docker.deb.amd64 \
	package.arhat-docker.deb.armv6 \
	package.arhat-docker.deb.armv7 \
	package.arhat-docker.deb.arm64

package.arhat-docker.rpm.amd64:
	sh scripts/package/package.sh $@

package.arhat-docker.rpm.armv7:
	sh scripts/package/package.sh $@

package.arhat-docker.rpm.arm64:
	sh scripts/package/package.sh $@

package.arhat-docker.rpm.all: \
	package.arhat-docker.rpm.amd64 \
	package.arhat-docker.rpm.armv7 \
	package.arhat-docker.rpm.arm64

package.arhat-docker.linux.all: \
	package.arhat-docker.deb.all \
	package.arhat-docker.rpm.all

#
# windows
#

package.arhat-docker.msi.amd64:
	sh scripts/package/package.sh $@

package.arhat-docker.msi.arm64:
	sh scripts/package/package.sh $@

package.arhat-docker.msi.all: \
	package.arhat-docker.msi.amd64 \
	package.arhat-docker.msi.arm64

package.arhat-docker.windows.all: \
	package.arhat-docker.msi.all

#
# darwin
#

package.arhat-docker.pkg.amd64:
	sh scripts/package/package.sh $@

package.arhat-docker.pkg.arm64:
	sh scripts/package/package.sh $@

package.arhat-docker.pkg.all: \
	package.arhat-docker.pkg.amd64 \
	package.arhat-docker.pkg.arm64

package.arhat-docker.darwin.all: \
	package.arhat-docker.pkg.all
