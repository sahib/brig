# Parcello Example

You should start the application with the following command:

```console
$ go run main.go
```

This example illustrates how to embed resource in Golang application. If you
want to enable dev mode, which enables editing content on a fly, you should set
the following environment variables before you start the application:

```console
$ export PARCELLO_DEV_ENABLED=1
$ export PARCELLO_RESOURCE_DIR=./public
```

or use tools like [direnv](https://direnv.net):

```console
$ direnv allow
```


