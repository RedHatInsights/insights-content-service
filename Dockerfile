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

FROM registry.access.redhat.com/ubi9/go-toolset:9.5-1737480393 AS builder

COPY . .

USER 0

# clone rules content repository and build the content service
RUN curl -ksL https://certs.corp.redhat.com/certs/2015-IT-Root-CA.pem -o /etc/pki/ca-trust/source/anchors/RH-IT-Root-CA.crt && \
    curl -ksL https://certs.corp.redhat.com/certs/2022-IT-Root-CA.pem -o /etc/pki/ca-trust/source/anchors/2022-IT-Root-CA.pem && \
    update-ca-trust && \
    umask 0022 && \
    git config --global --add safe.directory /opt/app-root/src && \
    make build && \
    chmod a+x insights-content-service && \
    ./update_rules_content.sh

FROM registry.access.redhat.com/ubi9/ubi-micro:latest

COPY --from=builder /opt/app-root/src/insights-content-service .
COPY --from=builder /opt/app-root/src/openapi.json /openapi/openapi.json
COPY --from=builder /opt/app-root/src/groups_config.yaml /groups/groups_config.yaml

# copy just the rule content instead of the whole ocp-rules repository
COPY --from=builder /opt/app-root/src/rules-content /rules-content

# copy the certificates from builder image
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /etc/pki /etc/pki

USER 1001

CMD ["/insights-content-service"]
