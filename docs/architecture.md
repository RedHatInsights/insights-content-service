---
layout: page
nav_order: 1
---

# Architecture

Content Service consists of three main parts:

1. A rules content parsing that reads the rules metadata from the defined repository, creating data
   structures.
1. A group configuration parser that reads a groups configuration file.
1. HTTP or HTTPS server that exposes REST API endpoints that can be used to read a single rule
   metadata content, a list of groups and a list of tags that belongs to a group.

---
**NOTE**

Detailed information about the exact format of data exposed via REST API is
available at
https://redhatinsights.github.io/insights-data-schemas/content_service.html

---
