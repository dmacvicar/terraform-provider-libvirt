# Alpine Container
This container is build from the `Alpine` distribution. This container is meant to serve as a slim down minimalistic 
container.

**NOTE**: If pulling from the `build` containers, make sure to pull from the `musl` container. Alpine uses `musl` instead
of `glibc`.


## All-in-One Container Usage
Building Container:

```console
docker build -f Dockerfile_all_in_one -t terraform:development-alpine . --build-arg VERSION=v0.5.2 --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```


## Build-Dependent Container Usage
Build appropriate `Build` container, in this case it would be `Dockerfile_musl`:

```cosnole
docker build -f Dockerfile_musl -t provider-libvirt:v0.5.2-musl . --build-arg VERSION=v0.5.2
```

Now build the main container:
```console
docker build -f Dockerfile_build_dependent -t terraform:development-alpine . --build-arg GO_OS=linux --build-arg GO_ARCH=amd64 --build-arg TERRAFORM_VERSION=0.11.14
```

