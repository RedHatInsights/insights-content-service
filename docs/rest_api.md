---
layout: page
nav_order: 4
---

# REST API

Content service provides information about its REST API scheme via the endpoint
`api/v1/openapi.json`. OpenAPI 3.0 is used to describe the schema; it can be
read by human and consumed by computers. 

For example, if content service is started locally, it is possible to read schema based on OpenAPI 3.0
specification by using the following command:

```shell
curl localhost:8080/api/v1/openapi.json
```

Please note that OpenAPI schema is accessible w/o the need to provide authorization tokens.
