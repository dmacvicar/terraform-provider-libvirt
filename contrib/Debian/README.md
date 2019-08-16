# Debian Based Container
This container is build from a `Debian` based distribution, in this case we are using `Ubuntu`. This container has
all the "bloat" of a normal Ubuntu instance. Thus you will find common packages already installed like `nano`, `gcc`, 
etc. This option would be good for people who want to test the functionality of the terraform libvirt provider.

**NOTE**: If pulling from the `build` containers, make sure to pull from the `glibc` container.


## All-in-One Container Usage
Building Container:

```console
docker build -f Dockerfile_all_in_one -t terraform:development-ubuntu . --build-arg VERSION=v0.5.2 --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```


## Build-Dependent Container Usage
Build appropriate `Build` container, in this case it would be `Dockerfile_glibc`:

```cosnole
docker build -f Dockerfile_glibc -t provider-libvirt:v0.5.2-glibc . --build-arg VERSION=v0.5.2
```

Now build the Main container:
```console
docker build -f Dockerfile_build_dependent -t terraform:development-ubuntu . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```

