FROM golang:1.10 as builder

ENV container docker

ADD . /go/src/github.com/argcv/hou

RUN cd /go/src/github.com/argcv/hou && bash ./build.sh

FROM scratch

# x509: failed to load system roots and no roots provided
COPY --from=builder  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /go/src/github.com/argcv/hou/hou /hou

WORKDIR /app

EXPOSE 6789

ENTRYPOINT ["/hou"]
