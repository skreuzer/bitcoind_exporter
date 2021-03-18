FROM quay.io/prometheus/busybox:latest
MAINTAINER Steven Kreuzer <skreuzer@FreeBSD.org>

COPY bitcoind_exporter /bin/bitcoind_exporter

EXPOSE 9960
ENTRYPOINT [ "/bin/bitcoind_exporter" ]
