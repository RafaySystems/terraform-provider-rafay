# terraform-provider-rafay
Rafay terraform provider 

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-hashicups
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples/resources/rafay_project/
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```
