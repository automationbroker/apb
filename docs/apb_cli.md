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
After modifying the APB as desired, we need to create a buildconfig so that an imagestream is populated in a namespace which `apb` can read from. We recommend using the `openshift` namespace for this since by default all imagestreams in `openshift` namespace are accessible to all authenticated users, but, `apb` works with any accesible namespace. This is documented in the [getting-started document]().
Once the imagestream exists in namespace `foo`. You can add a new registry adapter with:
```
apb registry add --type local_openshift --namespaces foo lo
```

##### Using a remote registry (DockerHub)
Once your image is pushed to an organization on Dockerhub, you can configure `apb` to check the registry for available APBs. If your images exist in organization `foo`, you can configure a new registry adapter with:
```
apb registry add --type dockerhub --org foo foo-dockerhub
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
apb bundle provision mediawiki-apb
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
Create binding out of secret `foo` and add it to Deployment Config `bar`
```bash
apb binding add foo bar
```

Our example APBs create secrets that match the name of the APB pod. So if you want to bind Postgresql APB to Mediawiki, you would first provision Postgresql (`apb bundle provision postgresql-apb`) and Mediawiki (`apb bundle provision mediawiki-apb`). Once they are done, you should see a secret named `bundle-<hash>` if you do `oc get secret`. Then find the name of the DeploymentConfig you want to bind to (`oc get dc`).If the DeploymentConfig is `mediawiki-1234` a binding command may look like
```
$ apb binding add bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40 mediawiki-1234
INFO Create a binding using secret [bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40] to app [mediawiki-1234]                                                                                     

Successfully created secret [bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40-creds] in namespace [apb].                                                                                          
Use the following command to attach the binding to your application:
oc set env dc/mediawiki-1234 --from=secret/bundle-772f6e70-3ee5-4fce-9c26-1dec57cc0c40-creds
```

Type the recommended command:
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
| --name, -n         | Name of clusterservicebroker to relist |


##### Examples
Force a relist of clusterservicebroker `foo-broker`
```bash
apb catalog relist --name foo-broker
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
apb config
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

---

