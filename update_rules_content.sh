#!/usr/bin/env bash
# Copyright 2020 Red Hat, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.



# Updates the ./rules-content directory with the latest rules to test with

set -exv

function clean_up() {
    rm -rf "$CLONE_TEMP_DIR"
}
trap clean_up EXIT

# Updated with every new ccx-rules-opc release.
CCX_RULES_OCP_TAG="2024.01.03"

RULES_REPO="https://gitlab.cee.redhat.com/ccx/ccx-rules-ocp.git"
SCRIPT_DIR="$(dirname "$(realpath "$0")")"
CONTENT_DIR="${SCRIPT_DIR}/rules-content"
TEST_RULE_CONTENT_DIR="${SCRIPT_DIR}/rules/test/content"

CLONE_TEMP_DIR="${SCRIPT_DIR}/.tmp"
RULES_CONTENT="${CLONE_TEMP_DIR}/content/"

if [ $# -ne 0 ]
then
    if [[ "$*" == *-test-rules-only* ]]
    then
        echo "Creating content dir with test rules"
        mkdir -p "${CONTENT_DIR}/external/"
        mkdir -p "${CONTENT_DIR}/internal/"
        mkdir -p "${CONTENT_DIR}/ocs/"
        cp -R "${SCRIPT_DIR}/rules" "${CONTENT_DIR}/external/."
        cp "${TEST_RULE_CONTENT_DIR}/config.yaml" "${CONTENT_DIR}/."
        exit 0
   fi
fi

echo "Attempting to clone repository into ${CLONE_TEMP_DIR}"
if ! git clone --depth=1 --branch "${CCX_RULES_OCP_TAG}" "${RULES_REPO}" "${CLONE_TEMP_DIR}"
then
    echo "Couldn't clone rules repository"
    exit $?
fi

if ! rm -rf "${CONTENT_DIR}"
then
    echo "Couldn't remove previous content"
    exit $?
fi

if ! mv "${RULES_CONTENT}" "${CONTENT_DIR}"
then
    echo "Couldn't move rules content from cloned repository"
    exit $?
fi

rm -rf "${CLONE_TEMP_DIR}"

cp "${TEST_RULE_CONTENT_DIR}/config.yaml" "${CONTENT_DIR}/external/rules/"

if [ $# -ne 0 ]
then
    if [[ "$*" == *-include-test-rules* ]]
    then
        echo "Including test rules in served content"
        cp -a "${TEST_RULE_CONTENT_DIR}/." "${CONTENT_DIR}/external/rules/"
    fi
fi

echo "${CONTENT_DIR} updated with latest rules"
exit 0
