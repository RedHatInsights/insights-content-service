#!/bin/bash
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

# this is improper - we need to start using tags in GitHub properly

set -exv

version=0.2

buildtime=$(date)
branch=$(git rev-parse --abbrev-ref HEAD)
commit=$(git rev-parse HEAD)

utils_version=$(go list -m github.com/RedHatInsights/insights-operator-utils | awk '{print $2}')

ocp_rules_version=$(grep "^CCX_RULES_OCP_TAG=\".*\"$" update_rules_content.sh | awk -F'CCX_RULES_OCP_TAG="|"' '{print $2}')

# Update ccx-rules-ocp
./update_rules_content.sh "$@"

build_flags="-v"

case "$*" in
(*-cover*) build_flags="-v -cover";;
esac

go build ${build_flags} -ldflags="-X 'main.BuildTime=$buildtime' -X 'main.BuildVersion=$version' -X 'main.BuildBranch=$branch' -X 'main.BuildCommit=$commit' -X 'main.UtilsVersion=$utils_version' -X 'main.OCPRulesVersion=$ocp_rules_version'"
exit $?
