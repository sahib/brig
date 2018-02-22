FROM golang
MAINTAINER sahib@online.de

ENV BRIG_USER alice@wonderland.lit/container
ENV BRIG_PATH /var/repo
RUN mkdir -p $BRIG_PATH

ENV BRIG_SOURCE /go/src/github.com/sahib/brig
COPY . $BRIG_SOURCE
WORKDIR $BRIG_SOURCE

RUN go install
RUN brig -x init $BRIG_USER

EXPOSE 6666
EXPOSE 4002

CMD ["brig", "-x", "--bind", "0.0.0.0", "-l", "stdout", "daemon", "launch"]
