# APB CLI Tool

`apb` is a tool for helping APB authors create, build, and publish
their APBs to container registries. It enforces best practices and takes
care of the details so they should be easy to deploy.

1. [Installation](#installing-the-apb-tool)
    * [Prerequisites](#prerequisites)
    * [Running from a container](#running-from-a-container)
    * [RPM Installation](#rpm-installation)
    * [Installing from source](#installing-from-source)
        * [Installing from source - Tito](#installing-from-source---tito)
    * [Test APB tooling](#test-apb-tooling)
    * [Access Permissions](#access-permissions)
1. [Typical Workflows](#typical-workflows)
    * [Creating and Testing APBs](#creating-and-testing-apbs)
    * [Using the internal OpenShift Registry](#using-the-internal-openshift-registry)
    * [Using a remote Registry](#using-a-remote-registry-dockerhub)
1. [APB Commands](#apb-commands)

## Installing the **_apb_** tool

#### Prerequisites

[Go](https://golang.org/) must be correctly installed and running on the system.

#### Running from a container

NOTE: You **must** configure your host networking to allow traffic between
your container and the minishift vm if you are using minishift. Execute [setup-network.sh](../scripts/setup-network.sh) to
setup necessary iptables rules.

Pull the container:
```bash
docker pull docker.io/ansibleplaybookbundle/apb-tools
```

There are three tags to choose from:
- **latest**: more stable, less frequent releases
- **nightly**: following upstream commits, installed from RPM
- **canary**: following upstream commits, installed from source build

Copy the [apb-docker-run.sh](https://raw.githubusercontent.com/ansibleplaybookbundle/ansible-playbook-bundle/master/scripts/apb-docker-run.sh) script into your `PATH` and
make sure it's executable:

```
cp $APB_CHECKOUT/scripts/apb-docker-run.sh $YOUR_PATH_DIR/apb && chmod +x $YOUR_PATH_DIR/apb
```

#### RPM Installation

For RHEL or CentOS 7:
```
su -c 'wget https://copr.fedorainfracloud.org/coprs/g/ansible-service-broker/ansible-service-broker-latest/repo/epel-7/group_ansible-service-broker-ansible-service-broker-latest-epel-7.repo -O /etc/yum.repos.d/ansible-service-broker.repo'

sudo yum -y install https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
sudo yum -y install apb
```


For Fedora 26, 27, 28+:
```
sudo dnf -y install dnf-plugins-core
sudo dnf -y copr enable @ansible-service-broker/ansible-service-broker-latest
sudo dnf -y install apb
```

#### Installing from source

##### Installing from source - Go

Install Go 1.8+.
```
sudo dnf install -y golang
```

Clone this repo into your `$GOPATH`
```
git clone https://github.com/automationbroker/apb.git
```

Install the `apb` tool into `$GOBIN`
```
cd apb && make install
```

##### Installing from source - Tito

Alternatively you can use [tito](http://github.com/dgoodwin/tito) to install.
```bash
tito build --test --rpm -i
```

#### Test APB Tooling
Run `apb help` to make sure the tool is installed correctly
```
$ apb help
Tool for working with Ansible Playbook Bundles

Usage:                                                                
  apb [command]
Available Commands:
  binding     Manage bindings
  broker      Interact with an Automation Broker instance
  bundle      Interact with ServiceBundles
  catalog     Interact with OpenShift Service Catalog
  completion  Generates shell completion scripts.
  config      Set tool defaults
  help        Help about any command
  registry    Configure registry adapters
  version     Get version

Flags:
      --config string   configuration file (default is $HOME/.apb)
  -h, --help            help for apb
  -v, --verbose         verbose output

Use "apb [command] --help" for more information about a command.
```

#### Access Permissions

The `apb` tool requires you to be logged in as a tokened cluster user (`system:admin`
is not sufficient because it does not have a token that can be used for the tool's authentication).
In addition, there are a number of `RoleBinding`s and `ClusterRoleBindings` that must
exist to permit the full breadth of the `apb` tool's functions.

The easiest option is to ensure the user has the `cluster-admin` `ClusterRoleBinding`.
**To be clear, this is effectively cluster root and should only be used in a development setting**.

```
oc adm policy add-cluster-role-to-user cluster-admin <user>
oc login -u <user>
```

If you would like a more strictly permissioned environment, we have an [Openshift Template](../templates/openshift-permissions.template.yaml)
available that can be applied with the following command:

`oc process -f templates/openshift-permissions.template.yaml -p USER=<your_desired_user> | oc create -f -`.

By default, the template will permission the `developer` user. If that is your user, it
is safe to leave off the `-p` flag, which overrides the default value. Obviously, this
command must be run by a user with sufficient permissions to create the various roles.
The `developer` account does not have such permissions. `oc login -u system:admin` should
be sufficient.

## Typical Workflows

#### Creating and Testing APBs
In version `2.0.0` and above, `apb` allows you as a developer to develop and test APB images without using an Ansible Service Broker. `apb` allows the user to configure registries to source APBs from and stores the list of images locally (in `~/.apb`). Once the images are found, `apb` allows the user to provision the APB directly for testing.


Initializing the APB is done with `ansible-galaxy`:
```
ansible-galaxy init --type apb <apb-name>
```

##### Using the internal OpenShift registry
After modifying the APB as desired, we need to create a buildconfig so that an imagestream is populated in a namespace which `apb` can read from. 

We recommend putting APB images into the `openshift` namespace since imagestreams in the `openshift` namespace are accessible to all authenticated users by default, but `apb` works with any accessible namespace. More information is available in [getting-started.md](getting-started.md).

You can make APB imagestreams in namespace `foo` accessible to the `apb` tool by adding a new registry adapter with:
```
apb registry add lo --type local_openshift --namespaces foo
```

##### Using a remote registry (DockerHub)
Once your image is pushed to an organization on Dockerhub, you can configure `apb` to check the registry for available APBs. If your images exist in organization `foo`, you can configure a new registry adapter with:
```
apb registry add foo-dockerhub --type dockerhub --org foo
```

To run the `provision` playbook:
```
apb bundle provision <apb_name> 
```

To view available APBs:
```
apb bundle list
```

## APB Commands
These are the top level commands with each subcommand documented under the parent:

[bundle](#bundle)

[broker](#broker)

[catalog](#catalog)

[binding](#bindina)

[completion](#completion)

[config](#config)

[help](#help)

[registry](#registry)

[version](#version)

---
### `bundle`

##### Description
Interact with Ansible Playbook Bundle images present in the `apb` tool

##### Usage
```bash
apb bundle [COMMAND] [OPTIONS]
```

##### Commands
| Subcommand  | Description |
| :---        | :---        |
| deprovision | Deprovision APB image |
| info        | Print info about APB image |
| list        | List available APB images |
| prepare     | Stamp APB metadata onto Dockerfile in base64 encoding |
| provision   | Provision APB images |
| test        | Test APB images |

##### Options

| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message |
| --kubeconfig, -k   | Path to kubeconfig to use |


##### Examples
Provision `mediawiki-apb` APB image
```bash
# Provision mediawiki-apb in the background
apb bundle provision mediawiki-apb

# Provision mediawiki-apb and follow APB logs
apb bundle provision mediawiki-apb --follow

# Provision mediawiki-apb using 'admin' sandbox-role
apb bundle provision mediawiki-apb --sandbox-role admin

# Deprovision mediawiki-apb without prompting for parameters and follow APB logs
apb bundle deprovision --skip-params --follow
```

---
### `binding`

##### Description
Manage bindings on an OpenShift cluster

##### Usage
```bash
apb binding [command]
```

##### Commands
| Subcommand | Description |
| :---       | :---        |
| add        | Add bind credentials to an application |

##### Options

| Option, shorthand      | Description |
| :---                   | :---        |
| --help, -h             | Show help message for binding |
| --namespace, -n        | Namespace of binding |

##### Examples
Create binding out of secret `foo-secret` and add it to Deployment Config `bar-dc`:
```bash
apb binding add foo-secret bar-dc
```

Our example APBs create secrets that match the name of the APB pod. 

To bind Postgresql APB to Mediawiki:
1. Provision Postgresql (`apb bundle provision postgresql-apb`)
1. Provision Mediawiki (`apb bundle provision mediawiki-apb`)
1. Wait for APBs to finish provisioning
1. Run `oc get secret`, look for a secret named `bundle-<hash>` that matches the hash from your Postgres APB run
1. Run `oc get dc` and identify the DeploymentConfig you want to add your bind secrets to
1. If the DeploymentConfig is named `mediawiki-1234` and the bundle hash is `772f6e70-[...]` a binding command may look like:
```
$ apb binding add bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40 mediawiki-1234

INFO Create a binding using secret [bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40] to app [mediawiki-1234]                                 
Successfully created secret [bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40-creds] in namespace [apb].                                      

Use the following command to attach the binding to your application:
oc set env dc/mediawiki-1234 --from=secret/bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40-creds
```

Type the recommended command to complete the binding:
```
$ oc set env dc/mediawiki-1234 --from=secret/bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40-creds
deploymentconfig "mediawiki-1234" updated
```

This will redeploy Mediawiki pod and you should see the full application backed by a Postgresql instance.

---
### `broker`

##### Description
Interact with Ansible Service Broker

Bootstrap and list available APBs in an Ansible Service Broker instance

##### Usage
```bash
apb broker [command]
```

##### Commands
| Subcommand | Description |
| :---       | :---        |
| bootstrap  | Bootstrap an Ansible Service Broker instance |
| catalog    | List available APBs in Anisble Service Broker catalog |

##### Options
| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message for broker |

##### Examples
Bootstrap an Ansible Service Broker instance using config values stored in ~/.apb/defaults.json
```bash
apb broker bootstrap
```

---
### `catalog`

##### Description
Interact with OpenShift Service Catalog. Force the Service Catalog to relist its available APBs from an Ansible Service Broker instance

##### Usage
```bash
apb catalog [COMMAND] [OPTIONS]
```

##### Commands
| Subcommand | Description |
| :---       | :---        |
| relist     | Force a relist of the OpenShift Service Catalog |

##### Options

| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message |
| --output, -o       | Display output as `yaml` or `json` |


##### Examples
Force a relist of the service broker specified by config values stored in ~/.apb/defaults.json
```bash
apb catalog relist
```

Force a relist of the service broker specified by config values stored in ~/.apb/defaults.json, printing output as JSON.
```bash
apb catalog relist -o json
```

---
### `config`

##### Description
Runs an interactive prompt to configure defaults for the `apb` tool

##### Usage
```bash
apb config [OPTIONS]
```

##### Options

| Option, shorthand   | Description |
| :---                | :---        |
| --help, -h          | Show help message |

##### Examples

Set new defaults for `apb`
```bash
$ apb config
Broker namespace [default: openshift-automation-service-broker]: 
Broker resource URL [default: /apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/]: 
Broker route name [default: openshift-automation-service-broker]: 
clusterservicebroker resource name [default: openshift-automation-service-broker]: 
# Broker route suffix values: 
# -------------------------------
# 3.9:   "ansible-service-broker"
# 3.10:  "ansible-service-broker"
# 3.11+: "osb"
Broker route suffix [default: osb]:                                     

Saving new configuration.... 
```

---
### `completion`

##### Description
Generates shell completion scripts. This gives completion scripts for bash and zsh.

##### Usage
```bash
apb completion [COMMAND] [OPTIONS]
```

##### Commands
| Subcommand | Description |
| :---       | :---        |
| bash       | Generate shell completion script for bash |
| zsh        | Generate shell completion script for zsh | 

##### Options

| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message |


##### Examples
Generate bash completion script
```bash
apb completion bash
```

---

### `help`

##### Description
Get help information for any command

##### Usage
```bash
apb help [COMMAND]
```

##### Examples

Get more information about the `apb broker` subcommand
```bash
apb help broker
```

---
### `registry`

##### Description
Add, list, or remove registry configurations from the `apb` tool. We support all available registry types available within the Ansible Service Broker

##### Usage
```bash
apb registry [COMMAND] [OPTIONS]
```

##### Commands
| Subcommand | Description |
| :---       | :---        |
| add        | Add a new registry adapter |
| list       | List the configured registry adapters |
| remove | Remove a registry adapter |

##### Options

| Option, shorthand   | Description |
| :---                | :---        |
| --help, -h          | Show help message |


##### Examples
Add a registry named `dockerhub` configured to use organization `dune` from Dockerhub
```bash
apb registry add dockerhub --org dune 
```

List configured registries
```bash
apb registry list
```

Remove registry named `dockerhub`
```bash
apb registry remove dockerhub
```

---
### `version`

##### Description
Displays current version of `apb` tool

##### Usage
```bash
apb version
```

---

