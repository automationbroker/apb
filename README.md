# sbcli

## Run setup_namespace.sh
Once logged into your OCP cluster, you can run our helper script to create a new namespace with an `apb` service account to be used when launching the bundle.
```
$ ./setup_namespace.sh test-project
```

## Provision a Service Bundle
To provision a Service Bundle into this project, run:
```
sbcli exec provision --name <bundle_name> -p test-project
```
