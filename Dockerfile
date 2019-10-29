# build lstvi builder
FROM docker.io/centos:7 AS builder
ENV GOPATH=/go
ENV TMPDIR=/go/src/github.com/paramite/lstvi

WORKDIR $TMPDIR
COPY . $TMPDIR/

RUN yum install -y epel-release && yum update -y && \
    yum install -y git golang && yum clean all && \
    go get -u github.com/golang/dep/... && /go/bin/dep ensure -v -vendor-only && \
    go build -o /tmp/lstvi main/main.go && \

# build lstvi runner
FROM docker.io/centos:7

RUN yum install epel-release -y && \ yum update -y && yum clean all

COPY --from=builder /tmp/lstvi /usr/bin/.

ENTRYPOINT ["/usr/bin/lstvi"]
