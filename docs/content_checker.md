---
layout: page
nav_order: 6
---

# Rule content checker

A utility for checking the rule content is currently included. 

**DISCLAIMER**: It may be moved elsewhere in the future.

It helps to ensure that:

* tags referenced in the rule content are defined in the group configuration
* rule content attributes and content files are not empty
* every group name is unique
* group tags are unique (within the group)

It is necessary to have the rule content available locally in order to run the
tool.

Once you have the rule content and the rule group configuration file, you can
run the checker tool using the following command. Make sure to replace the
placeholders with actual paths. The content directory must be the one containing
the `config.yaml` file and the `external` directory with content for external
rules. Other rules are not being checked by this tool at the moment.

```shell
go run ./checker/ -config GROUP_CONFIG_YAML_PATH -content CONTENT_DIR_PATH
```

After running this command, you should see a report for the given group
configuration file and rule content directory in the terminal.

After checking each error code, a summary is printed containing its tags and a
list of groups to which the individual tags belong. Tags that do not belong to
any defined group are reported as an error and will not be included in this
summary.
