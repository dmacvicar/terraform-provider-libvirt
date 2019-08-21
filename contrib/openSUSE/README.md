# openSUSE Container
This container is based on openSUSE's `Tubmleweed`. `Tumbleweed` was chosen over `Leap` due to its multi-arch
support.

**NOTE**: If pulling from the `build` containers, make sure to pull from the `glibc` container.


## All-in-One Container Usage
Building Container:

```console
docker build -f Dockerfile_all_in_one -t terraform:development-tumbleweed . --build-arg VERSION=v0.5.2 --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```


## Build-Dependent Container Usage
Build appropriate `Build` container, in this case it would be `Dockerfile_glibc`:

```cosnole
docker build -f Dockerfile_glibc -t provider-libvirt:v0.5.2-glibc . --build-arg VERSION=v0.5.2
```

Now build the main container:
```console
docker build -f Dockerfile_build_dependent -t terraform:development-tumbleweed . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```

