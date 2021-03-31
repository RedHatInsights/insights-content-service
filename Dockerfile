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

FROM registry.redhat.io/rhel8/go-toolset:1.14 AS builder

COPY . .

ARG GITHUB_API_TOKEN

ENV RULES_CONTENT_DIR=/rules-content \
    RULES_REPO=https://github.com/RedHatInsights/ccx-rules-ocp/ \
    GIT_ASKPASS=/tmp/git-askpass.sh \
    CCX_RULES_OCP_VERSION=2021.03.31

USER 0

# clone rules content repository and build the content service
RUN umask 0022 && \
    mkdir -p $RULES_CONTENT_DIR && \
    echo "echo $GITHUB_API_TOKEN" > $GIT_ASKPASS && \
    chmod +x $GIT_ASKPASS && \
    git -C $RULES_CONTENT_DIR clone --depth=1 --branch $CCX_RULES_OCP_VERSION $RULES_REPO $RULES_CONTENT_DIR && \
    make build && \
    chmod a+x insights-content-service

FROM registry.redhat.io/ubi8-minimal

COPY --from=builder /opt/app-root/src/insights-content-service .
COPY --from=builder /opt/app-root/src/openapi.json /openapi/openapi.json
COPY --from=builder /opt/app-root/src/groups_config.yaml /groups/groups_config.yaml
# copy just the rule content instead of the whole ocp-rules repository
COPY --from=builder /rules-content/content/ /rules-content
# copy tutorial/fake rule to external rules to be hit by all reports
COPY rules/tutorial/content/ /rules-content/external/rules

RUN curl -L -o /usr/bin/haberdasher \
https://github.com/RedHatInsights/haberdasher/releases/download/v0.1.3/haberdasher_linux_amd64 && \
chmod 755 /usr/bin/haberdasher

USER 1001

ENTRYPOINT ["/usr/bin/haberdasher"]

CMD ["/insights-content-service"]
