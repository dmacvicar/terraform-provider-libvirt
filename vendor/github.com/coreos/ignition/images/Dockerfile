# Used in the Jenkins environment

FROM golang:_GOVERSION_

RUN echo "deb http://deb.debian.org/debian stretch main" >> /etc/apt/sources.list

# gcc for cgo
RUN apt-get update && apt-get install --force-yes -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
		gcc-aarch64-linux-gnu \
		libc6-dev-arm64-cross \
		libblkid-dev \
	&& rm -rf /var/lib/apt/lists/*
