# KubeKit Documentation

To access the KubeKit Documentation go to https://github.com/pages/kubekit/kubekit/. This document only explain how to edit the documentation.

Most of the documents were created in Markdown then exported to HTML. To edit the documents it's recommended to use a Markdown editor such as [Typora](https://typora.io/) or [VSCode](https://code.visualstudio.com/) with a Markdown preview plugin like [Mermaid Preview](https://marketplace.visualstudio.com/items?itemName=vstirbu.vscode-mermaid-preview) but for small changes any editor is fine.

## Technical Documentation

The flowcharts were done using [Mermaid](https://mermaidjs.github.io/flowchart.html), edited and exported to HTML with [Typora](https://typora.io/).

There is one flowchart per document, each flowchart cover a kubekit command, i.e. commands `init` used as `kubekit init --cluster name`

For more information about Mermaid go to:

* [Mermaid](https://mermaidjs.github.io/)
* [Mermaid Flowcharts](https://mermaidjs.github.io/flowchart.html)

To export the Mermaid Markdown to HTML using Typora, open and edit the Markdown document (i.e. `init.md`) then export it to HTML with the menu option File -> Export -> HTML.

Verify the changes with a browser going to the generated html file using `file://`, i.e. [file:///Users/ja186051/Workspace/src/github.com/kubekit/kubekit/docs/init.html](file:///Users/ja186051/Workspace/src/github.com/kubekit/kubekit/docs/init.html).  After pushing the changes to master branch, double check going to the github page, i.e. https://github.com/pages/kubekit/kubekit/init.html.
