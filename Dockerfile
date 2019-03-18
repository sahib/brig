FROM golang
MAINTAINER sahib@online.de

# Most test cases can use the pre-defined BRIG_PATH.
ENV BRIG_PATH /var/repo
RUN mkdir -p $BRIG_PATH
ENV BRIG_USER="charlie@wald.de/container"

# Build the brig binary:
ENV BRIG_SOURCE /go/src/github.com/sahib/brig
ENV BRIG_BINARY_PATH /usr/bin/brig
COPY . $BRIG_SOURCE
WORKDIR $BRIG_SOURCE
RUN make

# Download IPFS, so the container can startup faster.
# (brig can also download the binary for you, but later)
RUN wget https://dist.ipfs.io/go-ipfs/v0.4.19/go-ipfs_v0.4.19_linux-amd64.tar.gz -O /tmp/ipfs.tar.gz
RUN tar xfv /tmp/ipfs.tar.gz -C /tmp
RUN cp /tmp/go-ipfs/ipfs /usr/bin

EXPOSE 6666
EXPOSE 4001

COPY scripts/docker-normal-startup.sh /bin/run.sh
CMD ["/bin/bash", "/bin/run.sh"]
