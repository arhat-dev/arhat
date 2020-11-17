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

test.pkg:
	sh scripts/test/unit.sh pkg

test.cmd:
	sh scripts/test/unit.sh cmd

test.tags:
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noconfhelper_pprof'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nosysinfo'

	# metrics (with peripheral metrics)
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nometrics'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nometrics noextension_peripheral'

	# extension codec
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nocodec_gogoprotobuf'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nocodec_stdjson'

	# storage drivers
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nostorage_general'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nostorage_sshfs'

	# network support
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nonethelper_stdnet'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nonethelper_piondtls'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='nonethelper_pipenet'

	# extensions
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noextension'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noextension_peripheral'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noextension_runtime'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noextension noextension_peripheral'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noextension noextension_runtime'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noextension noextension_peripheral noextension_runtime'

	# exec try support
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry_tar'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry_test'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry_tar noexectry_test'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry noexectry_tar'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry noexectry_test'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noexectry noexectry_tar noexectry_test'

	# connectivity methods
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_mqtt'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_coap'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_grpc'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_grpc noclient_mqtt'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_coap noclient_grpc'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_coap noclient_mqtt'
	CGO_ENABLED=1 $(MAKE) arhat TAGS='noclient_mqtt noclient_grpc noclient_coap'
