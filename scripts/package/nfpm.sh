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

set -e

. scripts/version.sh

BUILD_DIR="build"
PACKAGE_DIR="${BUILD_DIR}/package"

mkdir -p ${PACKAGE_DIR}

NFPM="$(command -v nfpm || printf "")"
command -v rpmbuild || NFPM=""

if [ -z "${NFPM}" ]; then
  NFPM="docker run --rm -v $(pwd):/tmp/pkg goreleaser/nfpm pkg"
  OUTPUT_DIR="/tmp/pkg/${PACKAGE_DIR}"
else
  NFPM="${NFPM} pkg"
  OUTPUT_DIR="${PACKAGE_DIR}"
fi

get_deb_arch() {
  case "${1}" in
    amd64)
      printf "amd64"
      ;;
    arm64)
      printf "arm64"
      ;;
    armv6)
      printf "armel"
      ;;
    armv7)
      printf "armhf"
      ;;
  esac
}

get_rpm_arch() {
  case "${1}" in
    amd64)
      printf "x86_64"
      ;;
    arm64)
      printf "aarch64"
      ;;
    armv6)
      printf ""
      ;;
    armv7)
      printf "armhfp"
      ;;
  esac
}

_write_nfpm_metadata() {
  file="${1}"

  cat > "${file}" <<EOF
platform: linux
section: default
priority: extra
epoch: 1
version: ${GIT_TAG:-dev}

maintainer: Arhat Dev Developers <dev@arhat.dev>
vendor: arhat.dev
homepage: https://arhat.dev
license: Apache 2.0

EOF
}

_do_package() {
  pkg_arch=$(get_${FORMAT}_arch ${ARCH})

  if [ -z "${pkg_arch}" ]; then
    echo "Arch ${ARCH} not supported"
    exit 1
  fi

  pkg_file="${COMP}.${pkg_arch}.${FORMAT}"

  config_file="${COMP}-${ARCH}.nfpm.yaml"
  cp "${PACKAGE_DIR}/${config_file}" "${PACKAGE_DIR}/${config_file}.${FORMAT}"

  config_file="${config_file}.${FORMAT}"
  cat >> "${PACKAGE_DIR}/${config_file}" <<EOF
arch: ${pkg_arch}
EOF
  ${NFPM} --config "${OUTPUT_DIR}/${config_file}" --target "${OUTPUT_DIR}/${pkg_file}"
  mv "${PACKAGE_DIR}/${pkg_file}" ${BUILD_DIR}/.
}

package_template_application_go() {
  config_file="${PACKAGE_DIR}/${COMP}-${ARCH}.config.yaml"
  cat > "${config_file}" <<EOF
TODO: add sample config here
EOF

  nfpm_config_file="${PACKAGE_DIR}/${COMP}-${ARCH}.nfpm.yaml"
  _write_nfpm_metadata "${nfpm_config_file}"

  # configure nfpm
  # https://nfpm.goreleaser.com/configuration/
  cat >> "${nfpm_config_file}" <<EOF
name: ${COMP}
description: <some description>
conflicts: []
# - <package name>

provides: []
# - <package name>

files:
  ${OUTPUT_DIR}/../../build/${COMP}.linux.${ARCH}: /usr/local/bin/${COMP}

config_files: {}
  # <local path>: <installation path>

scripts: {}
  # postinstall: ${OUTPUT_DIR}/../../scripts/package/nfpm-${COMP}-post-install.sh
  # preinstall: ${OUTPUT_DIR}/../../scripts/package/nfpm-${COMP}-pre-install.sh
  # preremove: ${OUTPUT_DIR}/../../scripts/package/nfpm-${COMP}-post-install.sh
  # postremove: ${OUTPUT_DIR}/../../scripts/package/nfpm-${COMP}-post-install.sh

overrides:
  rpm:
    depends: []
    # - <package name>
  deb:
    depends: []
    # - <package name>
EOF
  _do_package
}

FORMAT="${1}"
COMP="${2}"
ARCH="${3}"

case "${FORMAT}" in
  deb|rpm)
    echo "ok"
    ;;
  *)
    echo "format ${FORMAT} not supported by nfpm"
    exit 1
esac

action="package_$(printf "%s" "${COMP}" | tr '-' '_')"

${action}
