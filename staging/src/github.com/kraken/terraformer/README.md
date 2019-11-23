# Terraformer

Go package to use Terraform using the Terraform library instead of the binary.

## Install

    go get -u github.com/kraken/terraformer

## Use

Create an instance of Terraformer using `New()` function.

    t, err := terraformer.New()
    if err != nil {
      return err
    }

Add the provisioners and providers using `AddProvisioner` and `AddProvider` respectivelly.

    t.AddProvider("aws", aws.Provider())

    t.AddProvisioner("chef", chef.Provisioner())

A **provider** is what will communicate with the infrastructure where you want to apply a change. Examples of providers are AWS, GCP, Azure, OpenStak, vSphere, Heroku, Kubernetes, among others. Check https://github.com/terraform-providers and https://www.terraform.io/docs/providers/index.html for large list of Terraform providers. Once you select the providers used in the code, they have to be imported in the Go code. For example, if you want to use the AWS provider:

    import "github.com/terraform-providers/terraform-provider-aws/aws"

The provider included in Terraformer by default is **null**.

The **provisioner** are used to execute scripts, transfer files or configuration management. A list of provisioners can be found in https://www.terraform.io/docs/provisioners/index.html. To avoid extra code and errors, no default provisioners are loaded by default.

If the infrastructure (cluster, VMs, containers) to modify exists it's necesary to update the current state, otherwise a new infrastructure will be created or you'll get an error.

To sent the current infrastructure state use `LoadState()`. LoadState() requires a Reader interface, means a instance that implement `Read()`, such as `os.File` or `bytes.Buffer`.

    f, err := os.Open("cluster.tfstate")
    if err != nil {
      return err
    }
    defer f.Close()

    err = t.LoadState( f )
    if err != nil {
      return err
    }

It's time now to let Terraformer knows the code or Terraform template to execute. This is done assigning the code to `Terraformer.Code` variable like this:

    t.Code = []byte(code)

The code may use variables, if so, make sure to load them using `Var(name, value)`:

    t.Var("count", count)

If you apply the changes now Terraformer will use the default logger which will print info, warn and error log levels using the following format:

    LEVEL [ yyyy/mm/dd hh:mm:ss ] TERRAFORMER: Log message

Examples:

    INFO  [ 2017/09/28 19:20:44 ] TERRAFORMER: Building AWS region structure
    WARN  [ 2017/09/28 19:20:40 ] TERRAFORMER: terraform: shadow graph disabled
    ERROR [ 2017/09/28 20:35:54 ] TERRAFORMER: root: eval: *terraform.EvalSequence, err: aws_instance.server.0: diffs didn't match during apply. This is a bug with Terraform and should be reported as a GitHub Issue.

If this logging is not what you want, there are several options. Terraformer accept a logger that implements `terraformer.Logger` interface defined in the file [logger.go](./logger.go#L13). After create the logger you can assign it to the terraformer using `Logger( logger )`.

Check the [Loggers](#loggers) section to view some options to create a logger.

Ready? Set? Go! Apply the change using `Apply( bool )`. Apply() receives a boolean, if **true** the infrastructure will be destroyed, make sure it is **false** to create a new infrastructure or modify an existing one.

    t.Apply(false)

When Apply is running you'll see a lot of information sent to Stdout. To reduce such information or to redirect it to somewhere (i.e. a file) use the `Logger()` method. **Still under construction**

When it is over, you may want to get the final state of the infrastructure. This is useful in case you want to modify it or terminate it. This is done with the `SaveState()` method which receives a Writer interface, means a instance that implements `Write()`, such as `os.File` or `bytes.Buffer`. In the previous example to load the state was used a file, in this example to save it, we'll use a buffer but make sure to save it to a file if you want the state to persist.

    var b bytes.Buffer
    t.GetState( b )
    fmt.Printf("Cluster state: %s", b)

## Loggers

Terraformer accept a logger that implements terraformer.Logger interface defined in [logger.go](./logger.go#L13). Here are 3 examples that you can use.

### terraformer.StdLogger

It's the same logger that is used by default but you can customize it by changing the output (i.e. a file or StdErr), log level or prefix. Use the function 

    NewLogger(w io.Writer, prefix string, level Level)

Example:

    l := terraformer.NewLogger(os.Stdout, "AWS", terraformer.LogLevelDebug)
    t.Logger(l)

### github.com/sirupsen/logrus.Logger

The logger from [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) implements terraformer.Logger and it is a powerfull logger, one of the most used in Go programs. Check the documentation to know all the features of logrus or this simple example:

    l := logrus.New()
    l.SetLevel(logrus.DebugLevel)
    l.Out = os.Stdout

    t.Logger(l)

This code creates log entries like these but using some light colors:

    DEBU[0001] root: eval: *terraform.EvalApply  
    INFO[0002] Building AWS region structure 
    WARN[0003] terraform: shadow graph disabled   

### github.com/johandry/log.Logger

Logrus allows us to create a TextFormater to define how to print the log entries. The johandry/log.Logger use a TextFormater to print log entries in pretty and friendly format, using a prefix and colors. It also easly configurable using Viper. With a Viper instance it gets the logger configuration, so it's a good logger to use with your CLI program.

    v := viper.New()

    // Usually you don't set these parameters like this, 
    // they are set from environment variables or binded to Cobra flags.
    v.Set(log.OutputKey, os.Stdout)
    v.Set(log.ForceColorsKey, true)
    v.Set(log.LevelKey, "debug")

    l := log.New(v)
    l.SetPrefix("AWS")

This code creates log entries like these but the timestamp and log level are coloured:

    [Sep 28 16:23:29.222] DEBUG AWS: Waiting for state to become: [running]
    [Sep 28 16:23:38.254] INFO  AWS: aws_instance.server.0: Still creating... (10s elapsed)

### Create a wrapper

If you have a logger (for example, `github.com/mitchellh/cli.Ui`) that does not implement `terraformer.Logger` you can create your own struct that implement the interface's functions. Inside these functions you may call the functions of the logger you want.

Check [`terraformer.StdLogger`](./logger.go#L138) as an example, it is a wrapper of `golang.org/pkg/log`.

## Troubleshooting

The Achilles heel of this code are the vendors. Unfortunatelly not all the Provisioner or Providers use the latest neither the same version of Terraform therefore we have to make sure that all the packages imported use the right version.

It's recommended to use a vendor manager such as `govendor` to make sure you have the right versions. If something fail, remove all your vendors and add them again.

[Dep](https://github.com/golang/dep) is a great tool but it have failed when trying to sync all the packages. It's still in development but once it's released we'll use it.

### Too many open files

If you get the error `too many open files` trying to update the vendors, increase the `ulimit -n` and run the update again.

    ulimit -n 5000

## Simple and Unpractical Example Code

This is simple example to create and destroy an instance in AWS (no practical at all) but in the repository [terraformer-examples](https://github.com/kraken/terraformer-examples) you'll find a lot more examples.

`main.go`:

    package main

    import (
      "log"
      "time"

      "github.com/terraform-providers/terraform-provider-aws/aws"
      "github.com/kraken/terraformer"
    )

    func main() {
      var c = 1

      var code = `
      variable "count" {
        description   = "Total instances to create"
        default       = 1
      }
      provider "aws" {
        region        = "us-west-2"
      }
      resource "aws_instance" "web" {
        instance_type = "t2.micro"
        ami           = "ami-6e1a0117"
        count         = "${var.count}"
      }`

      t, err := terraformer.New()
      if err != nil {
        log.Fatalf("Fail to create an empty Terraformer instance")
      }

      t.AddProvider("aws", aws.Provider())

      t.Code = []byte(code)

      t.Var("count", c)

      t.Apply(false)

      // Here you'll do something with that instance. Meanwhile you think what to
      // do, check it is initializing in AWS
      time.Sleep(30 * time.Second)

      t.Apply(true)
    }

Before build this code, make sure you have all the vendors:

    govendor init
    govendor list -no-status +missing | xargs -n1 go get -u
  	govendor add +external
    go build -o aws_instance .

Now execute it and check what's happening in your AWS console:

    export AWS_ACCESS_KEY_ID={my_aws_access_key}
    export AWS_SECRET_ACCESS_KEY={my_aws_secret_key}
    ./aws_instance
