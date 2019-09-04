# Community Driven Docker Examples
These docker containers are meant to serve as an isolated development environment. The most common use case for these 
containers are to run your terraform environment in an isolate container talking to a `remote` libvirt system. 

Please refer to the distro's `README.md` for specific instructions.

Additionally, you will find a `Build` folder along side the distro folders. The docker containers within the `Build`
folder will compile the terraform libvirt provider for you. Please refer to the `Build` folder's `README.md` for more 
information and instructions.

## Table of Content

**Distro Containers**
- [Alpine Containers](Alpine/)
- [Debian Containers](Debian/)
- [openSUSE Containers](openSUSE/)

**Build Containers**
- [Build Containers](Build/)


## Quickstart
 1. Grab the all-in-one container for the distro you want, in this case we are using `openSUSE`
     ```console
     git clone https://github.com/dmacvicar/terraform-provider-libvirt.git; cd ./terraform-provider-libvirt/contrib/openSUSE/
     ```
 2. Build the all-in-one container
    ```console
    docker build -f Dockerfile_all_in_one -t terraform:development-tumbleweed . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64
    ```
 3. Run the docker container
    ```console
    docker run -it terraform:development-tumbleweed /bin/bash
    ```

## General Usage
There are two types of containers you'll find for each distro, `All-in-One` and `Build-Dependent`.

### All-in-One v.s Build-Dependent Containers
The `All-in-One` container is a single Dockerfile that you can build and run. The Dockerfile takes advantage of 
Docker's multi-stage build functionality. It allows you to build a binary object in a "sub container" and then 
access the binary object in your "main container". This Dockerfile is useful if you want to get up and running 
quickly, but might lead to confusion down the road if plan on using multiple versions of the terraform libvirt provider.

The `Build-Dependent` container relies on another Dockerfile which contains the build of the plugin. To use this 
container you need to first build the appropriate build Dockerfile and tag it correctly. Once that is done, you are
able to reference it inside the `Build-Dependent`'s Dockerfile. This design is useful when you are dealing with 
multiple versions of the terraform libvirt provider such as, `0.5.2`, which works with Terraform `0.11.X` 
and the latest branch, `master`, which currently works with Terraform `0.12.x`.

### Build Argument Reference
The following arguments are supported:

- `VERSION` - (Required) A name of a branch or tag of the terraform libvirt plugin to build
- `TERRAFORM_VERSION` - (Optional) - A tag of the official docker terraform image to pull from
- `GO_ARCH` - (Required) The GO name of your given architecture
- `GO_OS` - (Required) The GO name of your given OS

### Examples
`Build` containers examples:

```console
docker build -f Dockerfile_glibc -t provider-libvirt:v0.5.2-glibc . --build-arg VERSION=v0.5.2
``` 

This command would checkout the tag `v0.5.2`.

```console
docker build -f Dockerfile_glibc -t provider-libvirt:master-glibc . --build-arg VERSION=master
```

This command would checkout the `master` branch.

`distro` containers examples:

```console
docker build -f Dockerfile_build_dependent -t terraform:development-tumbleweed . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```

This command builds a distro container, tags it as `terraform:development-tumbleweed`, sets the `GO_OS` to linux and
`GO_ARCH` to amd64 and sets the terraform version to `0.11.14`.

```console
docker build -f Dockerfile_build_dependent -t terraform:development-debian . --build-arg GO_OS=linux --build-arg GO_ARCH=s390x
```

This command builds a distro container, tags it as `terraform:development-debian`, sets the `GO_OS` to linux and
`GO_ARCH` to s390x and will use the default value of `0.12.0` for terraform.

### Running on non-supported Terraform Architectures 
Terraform currently does support other architectures other then `amd64`, thus running on other architectures like 
`s390x` can be troublesome. 

The docker containers only needs a slight modification to run on `s390x`. **Note**: The `Build` containers will run
on any architectures that support GO and should not need modification.

In the distro containers you should see a line like:

```dockerfile
# Grab the Terraform binary
FROM hashicorp/terraform:$TERRAFORM_VERSION AS terraform
``` 

Currently, that is pulling in the official Terraform docker container. To get it to work on your desired architecture
you need to build the Terraform binary yourself. You can do that by building the Dockerfile below:

```dockerfile
FROM golang:alpine

ARG TERRAFORM_VERSION
ENV TERRAFORM_VERSION=$TERRAFORM_VERSION

RUN apk add --update git bash openssh

ENV TF_DEV=true
ENV TF_RELEASE=true

WORKDIR $GOPATH/src/github.com/hashicorp/terraform
RUN git clone https://github.com/hashicorp/terraform.git ./ && \
    git checkout v${TERRAFORM_VERSION} && \
    /bin/bash scripts/build.sh

WORKDIR $GOPATH
ENTRYPOINT ["terraform"]
``` 

With this Dockerfile built, you now need to swap the `FROM hashicorp/terraform:$TERRAFORM_VERSION AS terraform` with 
your images name and tag.

**Note**: Even if you get the terraform binary built for your respective architecture, you might need to build other
providers you utilize in your terraform files. As the default providers are not built for unsupported architectures.

### Tips and Tricks
- The use of Docker Volumes helps transfer Terraform config files back and forth between your local system and the 
docker container.

- Most `remote` libvirt systems require SSH Key auth. To generate a new SSH Key in Dockerfile use the following code:
    ```dockerfile
    # Make SSH Key
    RUN mkdir -p /root/.ssh/
    RUN touch /root/.ssh/id_rsa
    RUN chmod 600 /root/.ssh/id_rsa
    RUN echo "Host *" > /root/.ssh/config && echo " StrictHostKeyChecking no" >> /root/.ssh/config
    ```