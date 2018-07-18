# Design

## Overview
An [Ansible Playbook Bundle (APB)](https://github.com/automationbroker/apb)
borrows several concepts from the [Nulecule](https://github.com/projectatomic/nulecule)
and the [Atomicapp](http://www.projectatomic.io/docs/atomicapp/) project, namely the concept of a short
lived container with the sole purpose of orchestrating the deployment of the intended application. For the case
of APB, this short lived container is the APB; a container with an Ansible runtime environment
plus any files required to assist in orchestration such as playbooks, roles, and extra dependencies.
Specification of an APB is intended to be lightweight, consisting of several named playbooks and a
metadata file to capture information such as parameters to pass into the application.

## Workflow
APB is broken up into the following steps.

  1. [Preparation](#preparation)
     * [APB init](#apb-initialization)
     * [Spec File](#spec-file)
     * [Actions](#actions) (provision, deprovision, bind, unbind)
  1. [Build](#build)
  1. [Deploy](#deploy)

### Preparation

#### APB Initialization
![Prepare](images/apb-prepare.png)
The first step to creating an APB is to run the `apb init` command, which will create the required skeleton directory structure, and a few required files (e.g. `apb.yml` spec file) for the APB.

The `metadata` field is optional and used when integrating with the origin service catalog.

APB's with no parameters would define the `parameters` field as follows:
```yml
parameters: []
```

*__NOTE:__* `bind_parameters` is an __EXPERIMENTAL__ feature.  The `bind_parameters` are not required and will not be used when executing the bind action by default.  Currently, running bind returns encoded values from the provision action.  In order to use `bind_parameters` and execute the bind playbook, ASB must enable the feature using `LAUNCH_APB_ON_BIND=true`.  In addition, Kubernetes has not yet implemented [asynchronous binding](https://github.com/kubernetes-incubator/service-catalog/issues/1209), but it is expected in the near future.  Until then, if the binding is not created quickly, the request will time out with unexpected results.

#### Actions
The following are the actions for an APB. At a minimum, an APB must implement the `provision` and `deprovision` actions.
 * provision.yml
   * Playbook called to handle installing application to the cluster
 * deprovision.yml
   * Playbook called to handle uninstalling
 * bind.yml
   * Playbook to grant access to another service to use this service, i.e. generates credentials
 * unbind.yml
   * Playbook to revoke access to this service
 * test.yml
   * Playbook to test the the APB is vaild. This is an optional action.

The required named playbooks correspond to methods defined by the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker). For example, when the
Ansible Service Broker needs to `provision` an APB it will execute the `provision.yml`.

After the required named playbooks have been generated, the files can be used directly to test management of the
application. A developer may want to work with this directory of files, make tweaks, run, repeat until they are
happy with the behavior. They can test the playbooks by invoking Ansible directly with the playbook and any
required variables.

### Build
The build step is responsible for building a container image from the named playbooks for distribution.
Packaging combines a base image containing an Ansible runtime with Ansible artifacts and any dependencies required
to run the playbooks. The result is a container image with an ENTRYPOINT set to take in several arguments, one of
which is the method to execute, such as provision, deprovision, etc.

![Package](images/apb-package.png)

### Deploy
Deploying an APB means invoking the container and passing in the name of the playbook to execute along with any
required variables. It’s possible to invoke the APB directly without going through the Ansible Service Broker.
Each APB is packaged so it’s ENTRYPOINT will invoke Ansible when run. The container is intended to be short-lived,
coming up to execute the Ansible playbook for managing the application then exiting.

In a typical APB deploy, the APB container will provision an application by running the `provision.yml` playbook
 which executes a deployment role. The deployment role is responsible for creating the OpenShift resources,
 perhaps through calling `oc create` commands or leveraging Ansible modules. The end result is that the APB runs
 Ansible to talk to OpenShift to orchestrate the provisioning of the intended application.

![Deploy](images/apb-deploy.png)
