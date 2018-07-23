# APB CLI Tool

`apb` is a tool for helping APB authors create, build, and publish
their APBs to container registries. It enforces best practices and takes
care of the details so they should be easy to deploy.

1. [Installation](#installing-the-apb-tool)
    * [Prerequisites](#prerequisites)
    * [Running from a container](#running-from-a-container)
    * [RPM Installation](#rpm-installation)
    * [Installing from source](#installing-from-source)
        * [Python/VirtualEnv](#installing-from-source---pythonvirtualenv)
        * [Installing from source - Tito](#installing-from-source---tito)
    * [Test APB tooling](#test-apb-tooling)
1. [Typical Workflows](#typical-workflows)
    * [Local Registry](#local-registry)
    * [Remote Registry](#remote-registry)
1. [APB Commands](#apb-commands)
    * [Creating APBs](#creating-apbs)
        * [init](#init)
        * [prepare](#prepare)
        * [build](#build)
        * [push](#push)
        * [test](#test)
    * [Broker Utilities](#broker-utilities)
        * [list](#list)
        * [bootstrap](#bootstrap)
        * [remove](#remove)
        * [relist](#relist)    
    * [Other](#other)
        * [help](#help)    


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


For Fedora 26 or Fedora 27:
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
  completion  Generates shell completion scripts.
  help        Help about any command
  registry    Configure registry adapters

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

#### Local Registry
In order to use the internal OpenShift Docker Registry to source APBs, you must have configured the Ansible Service Broker to use the `local_openshift` type registry adapter. Please see the [config](https://github.com/openshift/ansible-service-broker/blob/master/docs/config.md#local-openshift-registry) section for more information.

```bash
apb init my-new-apb
cd my-new-apb
apb build
apb push
apb list
```

If you are using a namespace other than the default `openshift` namespace to host your APBs then you can use the following command:
```
apb push --namespace <namespace>
```

#### Remote Registry
Ansible Service Broker can also be [configured](https://github.com/openshift/ansible-service-broker/blob/master/docs/config.md#dockerhub-registry) to use a remote registry and org such as [docker.io/ansibleplaybookbundle](https://hub.docker.com/u/ansibleplaybookbundle/) or your own personal account.  In order to use this for developing APBs, you can build and push to your remote registry and then `bootstrap` to reload your APBs.

```bash
apb init my-new-apb
cd my-new-apb
apb build --tag docker.io/my-org/my-new-apb
docker push docker.io/my-org/my-new-apb
apb bootstrap
apb list
```

## APB Commands
[Creating APBs](#creating-apbs)
* [init](#init)
* [prepare](#prepare)
* [build](#build)
* [push](#push)
* [test](#test)
    
[Broker Utilities](#broker-utilities)
* [list](#list)
* [bootstrap](#bootstrap)
* [remove](#remove)
* [relist](#relist)    

[Other](#other)
* [help](#help)    

<a id="creating-apbs"></a>

---
### `binding`

##### Description
Manage bindings on an OpenShift cluster

##### Usage
```bash
apb binding [command]
```

##### Commands
_add_: Add bind credentials to an application

##### Options

| Option, shorthand      | Description |
| :---                   | :---        |
| --help, -h             | Show help message for binding |
| --namespace, -n        | Namespace of binding |

##### Examples
Create binding out of secret `foo` and add it to Deployment Config `bar`
```bash
apb binding add foo bar
```

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
_bootstrap_: Bootstrap an Ansible Service Broker instance
_catalog_: List available APBs in Anisble Service Broker catalog

##### Options
| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message for broker |
| --name, -n         | Name of Ansible Service Broker instance |

##### Examples
Bootstrap an Ansible Service Broker instance with the name `openshift-ansible-service-broker`
```bash
apb broker bootstrap --name openshift-ansible-service-broker
```

List available APBs in an Ansible Service Broker instance with the name `foo-broker`
```bash
apb broker catalog --name foo-broker
```

---
### `bundle`

##### Description
Interact with Ansible Playbook Bundle images present in the `apb` tool

##### Usage
```bash
apb bundle [COMMAND] [OPTIONS]
```

##### Commands
_deprovision_: Deprovision APB image
_info_: Print info about APB image
_list_: List available APB images
_prepare_: Stamp APB metadata onto Dockerfile in base64 encoding
_provision_: Provision APB images

##### Options

| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message |
| --kubeconfig, -k   | Path to kubeconfig to use |


##### Examples
Provision `mediawiki-apb` APB image
```bash
apb bundle provision mediawiki-apb
```

---
### `catalog`

##### Description
Interact with OpenShift Service Catalog. Force the Service Catalog to relist it's available APBs from an Ansible Service Broker instance

##### Usage
```bash
apb catalog [COMMAND] [OPTIONS]
```

##### Commands
_relist_: Force a relist of the OpenShift Service Catalog

##### Options

| Option, shorthand  | Description |
| :---               | :---        |
| --help, -h         | Show help message |
| --name, -n         | Name of clusterservicebroker to relist |


##### Examples
Force a relist of `foo-broker`
```bash
apb catalog relist --name foo-broker
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
_bash_: Generate shell completion script for bash
_zsh_: Generate shell completion script for zsh

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
_add_: Add a new registry adapter
_list_: List the configured registry adapters
_remove_: Remove a registry adapter

##### Options

| Option, shorthand   | Description |
| :---                | :---        |
| --help, -h          | Show help message |


##### Examples
Add a registry named `dockerhub` configured to use organization `dune` from Dockerhub
```bash
apb registry add --org dune dockerhub
```

List configured registries
```bash
apb registry list
```

Remove registry `dockerhub`
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
