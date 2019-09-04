# Debian Based Container
This container is built from a `Debian` based distribution, in this case we are using `Ubuntu`.

**NOTE**: If pulling from the `build` containers, make sure to pull from the `glibc` container.


## All-in-One Container Build Example
```console
docker build -f Dockerfile_all_in_one -t terraform:development-ubuntu . --build-arg VERSION=v0.5.2 --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```


## Build-Dependent Container Build Example
Build appropriate `Build` container, in this case it would be `Dockerfile_glibc`:

```cosnole
docker build -f Dockerfile_glibc -t provider-libvirt:v0.5.2-glibc . --build-arg VERSION=v0.5.2
```

Now build the main container:
```console
docker build -f Dockerfile_build_dependent -t terraform:development-ubuntu . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```

