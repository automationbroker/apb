# apb

## Getting Started
### Create a namespace
Once logged into your OCP cluster, you must first create a project:
```
$ oc new-project test-project
```

### Add a registry
First add the default registry (ansibleplaybookbundle on Dockerhub):
```
$ apb registry add
```
Or to add a custom registry:
``` 
$ apb registry add --name docker --org ansibleplaybookbundle --type dockerhub
```

Then you can list your registries:
```
$ apb registry list
Found registries already in config:                                   
 NAME       TYPE          ORG                       URL               
 ------ -+- --------- -+- --------------------- -+- ---------         
 docker  |  dockerhub  |  ansibleplaybookbundle  |  docker.io
```

### List available Service Bundles

Once a registry is configured, run `apb bundle list`:
```
$ apb bundle list
Found specs already in registry: [docker]                                                                                                    
 BUNDLE                    IMAGE                                                                   REGISTRY                                  
 --------------------- -+- ------------------------------------------------------------------- -+- --------                                  
 blankvm-apb            |  docker.io/ansibleplaybookbundle/virtualmachines-apb:latest           |  docker                                    
 prometheus-apb         |  docker.io/ansibleplaybookbundle/prometheus-apb:latest                |  docker                                    
 dynamic-apb            |  docker.io/ansibleplaybookbundle/dynamic-apb:latest                   |  docker                                    
 eclipse-che-apb        |  docker.io/ansibleplaybookbundle/eclipse-che-apb:latest               |  docker                                    
 etherpad-apb           |  docker.io/ansibleplaybookbundle/etherpad-apb:latest                  |  docker                                    
 pyzip-demo-db-apb      |  docker.io/ansibleplaybookbundle/pyzip-demo-db-apb:latest             |  docker
```
This command loads and caches Service Bundle specs (metadata) from newly added registries. Cached specs can be updated with `apb bundle list --refresh`

### Provision a Service Bundle
Provision any of the listed Service Bundles like so:
```
$ apb bundle provision <bundle_name> -n my-project
```

To provision a Service Bundle with the `admin` ClusterRole, run:
```
$ apb bundle provision <bundle_name> -n my-project -r admin
```


## Tips
### Enabling tab completion for apb
`apb` supports command tab completion for `bash` and `zsh`:
```
$ source <(apb completion bash) # load bash completion into session
$ source <(apb completion zsh) # load zsh completion into session
```

## Troubleshooting
### Using apb with openshift.io
In the starter tier of openshift.io, the secret quota is 20 (at the time of this writing). Since each `apb bundle provision` action creates a serviceaccount with an API token residing in a secret, it's possible to hit this limit after a few provisions. Cleaning up `bundle-*****-[..]` serviceaccounts will help resolve this. You can check if you're hitting this limit with `oc get quota object-counts -o yaml`
