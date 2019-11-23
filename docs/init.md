# Diagram Flow

```mermaid
graph TD
  %% Init command
  %% kubekit init --cluster <name>
  title[<b>Init</b> command<br/><br/><i>kubekit init --cluster cluster-name</i>]
  title-->start
  style title fill:#FFF,stroke:#FFF
  linkStyle 0 stroke:#FFF,stroke-width:0;

  start((Start)) --> f1("cmd.initCluster( clusterName, format )")
  f1 --> f1.1("kluster.Unique(clusterName)")
  f1.1 -.-> f1.2("cmd.getEnvConfig()")
  f1.2 -.-> f1.3("kluster.NewPath( ... )")
  f1.3 -.-> f1.4("kluster.New( ... )")

  f1.4 --> new1["Kluster{...}"]
  subgraph New
  new1 -.cluster.-> f1.4.1("provisioner.SupportedPlatforms()")
  f1.4.1 -.platforms.-> loop["range platforms"]
  loop -.name,platform.-> f1.4.2("platform.Config()")
  f1.4.2 -."cluster.Platforms[name]".-> new2["&State{...}"]
  new2 -."cluster.State[name]".-> assign["platform"]
  assign -."cluster.provisioner[name]".-> endloop(("fa:fa-sync"))
  endloop -.-> loop
  endloop -.-> f1.4.3("configurator.DefaultConfig()")
  f1.4.3 -.cluster.Config.-> r((*Kluster))
  end

  r -.cluster.-> f1.5("cluster.Save()<br><i>Description in 'Provision' command</i>")
  f1.5 -.-> e((End))

  classDef point fill:#ccf,stroke:#333,stroke-width:4px;
  class endloop point

  click f1.5 "provision.html" "cluster.Save() explained in provision command"
```
