---
name: Bug report
about: Create a report to help us improve KubeKit
title: "[BUG]"
labels: bug

---

<!--
A great way to contribute to the project is to send a detailed report when you
encounter an issue. We always appreciate a well-written, thorough bug report,
and will thank you for it!
The GitHub issue tracker is for bug reports and feature requests. General support can be found at Slack - kubekit-workspace.slack.com #development channel
-->

**Community Note**

* Please, make sure that we do not have any duplicate issues already open. Search the issue list for this
repository and if there is a duplicate, please close your issue and add a comment to the existing issue instead, you can use the "subscribe" button to get notified on updates
* Please vote on this issue by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue to help the community and maintainers prioritize this request
* Please do not leave "+1" or "me too" comments, they generate extra noise for issue followers and do not help prioritize the request
* If you are interested in working on this issue or have submitted a pull request, please leave a comment
* If you have ways to reproduce the issue or have additional information that may help
resolving the issue, please leave a comment.

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior. Example:

1. Get the modules '...'
2. Create the cluster with the following settings '....'
3. Apply the changes on the platform '....'
4. Got error '...'

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots or logs**
If applicable, add screenshots to help explain your problem.
For lengthy log-files, consider posting them as a [gist](https://gist.github.com) or use services line Dropbox and share a link to the Zip file.
Don't forget to remove sensitive data from your logfiles before posting. You can replace those parts with "REDACTED".

**Settings**
List the settings used to build, modify or destroy the cluster. Include the version and the cluster config file, make sure to exclude sensitive data.

**Versions:**
Execute `kubekit version --verbose` and insert the output:

```text
KubeKit v2.1.0-dev
Kubernetes version: 1.15.5
Docker version: 19.03.1
etcd version: v3.4.1
```

**Environment variables:**
Execute `env | grep KUBEKIT` and insert the output, include the OS and version:

* OS: [e.g. macOS]
* Version: [e.g. 10.14.6]
* Variables:

 ```bash
 KUBEKIT_EDITOR=/usr/local/bin/code
 ```

**Additional context:**
Add any other context about the problem here.

**References:**
List any other GitHub issues (open or closed) or Pull Requests that should be linked to this bug. For example:

* #6017
