FROM golang
MAINTAINER sahib@online.de

# Most test cases can use the pre-defined BRIG_PATH.
ENV BRIG_PATH /var/repo
RUN mkdir -p $BRIG_PATH

# Build the brig binary:
ENV BRIG_SOURCE /go/src/github.com/sahib/brig
COPY . $BRIG_SOURCE
WORKDIR $BRIG_SOURCE
RUN make

EXPOSE 6666
EXPOSE 4001

COPY scripts/docker-normal-startup.sh /bin/run.sh
CMD ["/bin/bash", "/bin/run.sh"]
