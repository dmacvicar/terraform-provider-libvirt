# Community Driven Docker Examples
These docker containers are meant to serve isolated develop/deployment environment. Each docker container has the 
terraform libvirt provider built and placed in custom plugins folder. Most of the use cases for these containers is to 
run your terraform environment in a isolate container talking to a `remote` libvirt system. 

Please refer to the distro's `README.md` for specific instructions.

Additionally, you will find a `Build` folder. The docker containers within this folder will build the terraform
libvirt provider for you. Please refer to the `README.md` for more information and instructions.

The docker containers found in this folder are community driven. Feel free to submit a PR with updates and fixes!


## Table of Content
- [Alpine Containers](Alpine/README.md)
- [Build Containers](Build/README.md)
- [Debian Conainers](Debian/README.md)
- [openSUSE Containers](openSUSE/README.md)


## General Usage
There are two types of containers you'll find for each distro, `All-in-One` and `Build-Dependent`.

### All-in-One v.s Build-Dependent Containers
The `All-in-One` container is a single Dockerfile that you can build and run. The Dockerfile take advantage of 
Docker's multi-stage build functionality. It allows you do build a binary or object in a "sub container", add access
the binary or object in your "main container". This Dockerfile is useful if you want to get up and running quick, but 
might lead to confusion down the road if plan on using multiple versions of the terraform libvirt provider.

The `Build-Dependent` container relies on another Dockerfile which contains the build of the plugin. To use this 
container you need to first build the appropriate build Dockerfile and tag it correctly. Once that is done, you are
able to reference it inside the `Build-Dependent`'s Dockerfile. This is design is useful when you are dealing with 
multiple versions of the terraform libvirt provider such as stable build, `0.5.2`, which works with Terraform `0.11.X` 
and the less stable, `master`, which works with Terraform `0.12.x`.

### Build Args
There are a couple build args to be aware of when building the various containers.

For the `Build` containers you only need to worry about the `VERSION` arg. The `VERSION` arg lets you build a specific
branch/tag of the terraform libvirt provider.



Examples:
```console
docker build -f Dockerfile_glibc -t provider-libvirt:v0.5.2-glibc . --build-arg VERSION=v0.5.2
``` 

This command would checkout the tag `v0.5.2`, thus building the terraform libvirt provider as `v0.5.2`.

```console
docker build -f Dockerfile_glibc -t provider-libvirt:master-glibc . --build-arg VERSION=master
```

This command would checkout the `master` branch, thus building the latest code in the master branch. 

For the `distro` containers there are three build args, `TERRAFORM_VERSION`, `GO_ARCH`, and `GO_OS`.

The build arg, `TERRAFORM_VERSION`, lets you select which terraform version you want to run. By default this is set to 
`0.12.0`, but can be overwritten by setting it as a Docker `build-arg`.

The `GO_ARCH` and `GO_OS` need to be passed in when building the container as they do `not` have defaults. The purpose
of these args are to allow multiple architectures to run these docker containers. If you are unsure of what your 
`GO_ARCH` and `GO_OS` should be please refer to 
[this](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63). For most users running on `amd64`, use 
`GO_OS=linux` and `GO_ARCH=amd64`.

If you are using `s390x`, change `GO_ARCH` to `GO_ARCH=s390x`.

Examples:
```console
docker build -f Dockerfile -t terraform:development-tumbleweed . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```

This command builds a distro container, tags it as `terraform:development-tumbleweed`, sets the `GO_OS` to linux and
`GO_ARCH` to amd64 and sets the terraform version to `0.11.14`.

```console
docker build -f Dockerfile -t terraform:development-debian . --build-arg GO_OS=linux --build-arg GO_ARCH=s390x
```

This command builds a distro container, tags it as `terraform:development-debian`, sets the `GO_OS` to linux and
`GO_ARCH` to s390x and will use the default value of `0.12.0` for terraform.

### Tips and Tricks
- The use of Docker Volumes helps transfer Terraform config files back and forth between your local system and the docker
container.

- Most `remote` libvirt systems require SSH Key auth. To generate a new SSH Key in Dockerfile use the following code:
    ```dockerfile
    # Make SSH Key
    RUN mkdir -p /root/.ssh/
    RUN touch /root/.ssh/id_rsa
    RUN chmod 600 /root/.ssh/id_rsa
    RUN echo "Host *" > /root/.ssh/config && echo " StrictHostKeyChecking no" >> /root/.ssh/config
    ```