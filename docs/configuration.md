---
layout: page
nav_order: 3
---

# Configuration
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

Content service expects a toml configuration file. Default one is `config.toml`
in working directory, but it can be overwritten by
`INSIGHTS_CONTENT_SERVICE_CONFIG_FILE` env var.

Also each key in config can be overwritten by corresponding env var. For example
if you have config like

```toml
[server]
address = ":8080"
api_prefix = "/api/v1/"
api_spec_file = "openapi.json"
```

and environment variables

```shell
INSIGHTS_CONTENT_SERVICE__SERVER__ADDRESS=":443"
INSIGHTS_CONTENT_SERVICE__SERVER__API_PREFIX="/api/v2/"
```

the actual server port will be 443 instead of 8080 and the API base endpoint
will be `/api/v2/` instead of `/api/v1/`.

It's very useful for deploying docker containers and keeping some of the
configuration outside the main configuration, like passwords and secret tokens.


## Server configuration

The HTTP server configuration is in section `[server]` in the
configuration file.

```toml
[server]
address = ":8080"
api_prefix = "/api/v1/"
api_spec_file = "openapi.json"
```

* `address` is the host and port which server should listen to
* `api_prefix` is the prefix for the REST API path
* `api_spec_file` is the location of a required OpenAPI specification file

## Groups configuration

The groups are defined in a YAML configuration file. You can find an example in
[groups_config.yaml](https://github.com/RedHatInsights/insights-content-service/blob/master/groups_config.yaml).

In order to define which groups configuration file is loaded by the service, you
should use the `[groups]` section in the configuration file:

```toml
[groups]
path = "groups_config.yaml"
```

Where `path` is the absolute or relative path to the groups configuration file.

## Static content configuration

This service parses the rules static content at startup. For that reason,
configuring the directory where the rules content is deployed is mandatory
within the configuration.

To define the path where the service will look up for rules content, you should
define the following:

```toml
[content]
path = "rules-content"
```

Where `path` can be the absolute or relative path to the rules content directory.

## Metrics configuration

Metrics configuration is in section `[metrics]` in config file

```toml
[metrics]
namespace = "mynamespace"
```

* `namespace` if defined, it is used as `Namespace` argument when creating all
  the Prometheus metrics exposed by this service.
  
## Logging configuration

Logging configuration is in section `[logging]` in config file

```toml
[logging]
debug = false
log_level = "info"
```

* `debug` if set to `true`, it will make the logs shown in console to be printed
  in a human readable format instead of JSON.
* `log_level` should be one of the following values: `debug`, `info`, `warn`,
  `warning`, `error` or `fatal`.
