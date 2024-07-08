#!/bin/bash
# Copyright 2022 Red Hat, Inc
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

set -exv


# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ccx-data-pipeline"  #  name of app-sre "application" folder this component lives in
REF_ENV="insights-production"
COMPONENT_NAME="insights-content-service"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/ccx-insights-content-service"
COMPONENTS="ccx-data-pipeline ccx-insights-results dvo-writer dvo-extractor insights-content-service ccx-smart-proxy ccx-mock-ams ccx-redis" # space-separated list of components to laod
COMPONENTS_W_RESOURCES="insights-content-service"  # component to keep
CACHE_FROM_LATEST_IMAGE="true"
DEPLOY_FRONTENDS="false"

export IQE_PLUGINS="ccx"
# Run all pipeline and ui tests
export IQE_MARKER_EXPRESSION="pipeline"
export IQE_FILTER_EXPRESSION=""
export IQE_REQUIREMENTS_PRIORITY=""
export IQE_TEST_IMPORTANCE=""
export IQE_CJI_TIMEOUT="30m"
export IQE_SELENIUM="false"
export IQE_ENV="ephemeral"
export IQE_ENV_VARS="DYNACONF_USER_PROVIDER__rbac_enabled=false"

changes_including_ocp_rules_version() {
    git log -1 HEAD . | grep "Bumped ccx-rules-ocp version"
}

create_junit_dummy_result() {
    mkdir -p 'artifacts'

    cat <<- EOF > 'artifacts/junit-dummy.xml'
	<?xml version="1.0" encoding="UTF-8"?>
	<testsuite tests="1">
	    <testcase classname="dummy" name="dummy-empty-test"/>
	</testsuite>
	EOF
}

function build_image() {
    source $CICD_ROOT/build.sh
}

function deploy_ephemeral() {
    source $CICD_ROOT/deploy_ephemeral_env.sh
}

function run_smoke_tests() {
    # component name needs to be re-export to match ClowdApp name (as bonfire requires for this)
    export COMPONENT_NAME="ccx-insights-content"
    source $CICD_ROOT/cji_smoke_test.sh
    source $CICD_ROOT/post_test_results.sh  # publish results in Ibutsu
}


# for ccx-rules-ocp version bump PRs skip pr_check.sh tests.
if changes_including_ocp_rules_version; then
    echo "Only ccx-rules-ocp version bump, exiting"
    create_junit_dummy_result
    exit 0
fi

# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh
echo "creating PR image"
build_image

echo "deploying to ephemeral"
deploy_ephemeral

echo "running PR smoke tests"
run_smoke_tests
