# sbcli

## Create a namespace
Once logged into your OCP cluster, you must first create a project:
```
$ oc new-project test-project
```

## Add a registry
```
$ sbcli registry add --name docker --org ansibleplaybookbundle --type dockerhub
```
or for the same result with defaults:
```
$ sbcli registry add
```
Then you can list your registries:
```
$ sbcli registry list
Found registries already in config:                                   
 NAME       TYPE          ORG                       URL               
 ------ -+- --------- -+- --------------------- -+- ---------         
 docker  |  dockerhub  |  ansibleplaybookbundle  |  docker.io
```

## List available Service Bundles
```
$ sbcli bundle list
```
The first time you run this command it will attempt to bootstrap any newly added registries. If it finds some already then it will used the cached list.

## Provision a Service Bundle
To provision a Service Bundle with the `admin` ClusterRole into this project, run:
```
sbcli bundle provision <bundle_name> -p test-project -r admin
```
