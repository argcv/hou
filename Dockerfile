FROM golang:1.18-alpine as builder

ENV container docker

ADD . /go/src/github.com/argcv/hou

RUN cd /go/src/github.com/argcv/hou && sh ./build.sh

#FROM scratch
FROM alpine

# x509: failed to load system roots and no roots provided
#COPY --from=builder  /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /go/src/github.com/argcv/hou/hou /hou

WORKDIR /app

EXPOSE 6789

ENTRYPOINT ["/hou"]
