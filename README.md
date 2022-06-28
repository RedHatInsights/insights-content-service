6

# Insights Content Service

[![GoDoc](https://godoc.org/github.com/RedHatInsights/insights-content-service?status.svg)](https://godoc.org/github.com/RedHatInsights/insights-content-service)
[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/insights-content-service/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-content-service)](https://goreportcard.com/report/github.com/RedHatInsights/insights-content-service)
[![Build Status](https://ci.ext.devshift.net/buildStatus/icon?job=RedHatInsights-insights-content-service-gh-build-master)](https://ci.ext.devshift.net/job/RedHatInsights-insights-content-service-gh-build-master/)
[![Build Status](https://travis-ci.org/RedHatInsights/insights-content-service.svg?branch=master)](https://travis-ci.org/RedHatInsights/insights-content-service)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/RedHatInsights/insights-content-service)
[![codecov](https://codecov.io/gh/RedHatInsights/insights-content-service/branch/master/graph/badge.svg)](https://codecov.io/gh/RedHatInsights/insights-content-service)
[![License](https://img.shields.io/badge/license-Apache-blue)](https://github.com/RedHatInsights/insights-content-service/blob/master/LICENSE)

Content service for Insights rules groups, tags, and content.

<!-- vim-markdown-toc GFM -->

* [Description](#description)
* [Documentation](#documentation)
* [Usage](#usage)
* [Contribution](#contribution)
* [Package manifest](#package-manifest)

<!-- vim-markdown-toc -->



## Description

Insights Content Service is a service that provides metadata information about rules that are being
consumed by Openshift Cluster Manager. That metadata information contains rule title, description,
remmediations, tags and also groups, that will be consumed primarily by
[Insights Results Smart Proxy](https://github.com/RedHatInsights/insights-results-smart-proxy).

## Documentation

Documentation is hosted on Github Pages <https://redhatinsights.github.io/insights-content-service/>.
Sources are located in [docs](https://github.com/RedHatInsights/insights-content-service/tree/master/docs).

## Usage

```
Usage:

    ./content-service [command]

The commands are:

    <EMPTY>             starts content service
    start-service       starts content service
    help                prints help
    print-help          prints help
    print-config        prints current configuration set by files & env variables
    print-groups        prints current groups configuration
    print-rules         prints current parsed rules
    print-parse-status  prints information about all rules that have been parsed
    print-version-info  prints version info

```

## Contribution

Please look into document [CONTRIBUTING.md](CONTRIBUTING.md) that contains all information about how to
contribute to this project.

## Package manifest

Package manifest is available at [docs/manifest.txt](docs/manifest.txt).
