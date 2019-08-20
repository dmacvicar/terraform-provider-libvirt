# Build Containers
These containers build the terraform libvirt provider. There are two Dockerfile due to the fact that some linux
systems use `musl` while others use `glibc`. These are just two different implementations of the `libc`, each having 
their benefits. "Most" systems use `glibc` but there are a couple linux distributions like `Alpine` that use `musl`. When
in doubt, use `glibc`.

These containers will mostly used as one of the stages in the multi-stage Dockefiles you'll find here. They also have
the benefit of storing the binary in the container, thus you could use a `docker copy` to grab the binary and use it 
on your local system. 

## General Usage
As stated before in the general [README](../README.md), these containers have build args. `VERSION` controls what 
build/version of the terraform libvirt provider you are going to compile. You can set `VERSION` to any branch or tag of
the repo. If you set `VERSION` to `v0.5.2` it would build that specific branch/tag of the project. Another common 
branch to set it to would be `master`. 


To build the two containers use these commands:

```console
docker build -f Dockerfile_glibc -t provider-libvirt:v0.5.2-glibc . --build-arg VERSION=v0.5.2
```

Which would build the `glibc` version with the code of the `v0.5.2` branch.

```console
docker build -f Dockerfile_musl -t provider-libvirt:master-musl . --build-arg VERSION=master
```

Which would build the `musl` version with the code of the `master` branch.