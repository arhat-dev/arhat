#!/bin/sh

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

common_go_test_env="GOOS=$(go env GOHOSTOS) GOARCH=$(go env GOHOSTARCH)"
common_go_test_flags="-mod=vendor -v -failfast -covermode=atomic"

_run_go_test() {
    env_vars="$1"
    flags="$2"
    tags="$3"
    dir="$4"

    set -ex
    eval "${common_go_test_env} ${env_vars} go test ${common_go_test_flags} -tags='${tags}' ${flags} ${dir}"
}

pkg() {
    _run_go_test \
        "CGO_ENABLED=1" \
        "-coverprofile=coverage.pkg.txt -coverpkg=./pkg/..." \
        "arhat netgo nokube nocloud containers_image_openpgp rt_docker" \
        "./pkg/..."
}

cmd() {
    _run_go_test \
        "CGO_ENABLED=0" \
        "-coverprofile=coverage.cmd.txt -coverpkg=./cmd/arhat-none/... " \
        "arhat netgo nokube nocloud rt_none nometrics nodevices" \
        "./cmd/arhat-none/..."

    _run_go_test \
        "CGO_ENABLED=0" \
        "-coverprofile=coverage.cmd.txt -coverpkg=./cmd/arhat-docker/... " \
        "arhat netgo nokube nocloud rt_docker nometrics nodevices" \
        "./cmd/arhat-docker/..."
}

$1
