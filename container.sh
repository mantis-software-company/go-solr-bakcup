#!/bin/sh
CONTAINER_EXEC=podman-remote
TARGET_HOST=$(printenv CONTAINER_HOST|sed 's-tcp://--g'|cut -f1 -d:)
REV=$(git describe --long --tags --match='v*' --dirty 2>/dev/null || echo dev)


$CONTAINER_EXEC build -f container/build.Containerfile -t mantissoftware/go-solr-backup:$REV . ||exit 1
$CONTAINER_EXEC tag mantissoftware/go-solr-backup:$REV mantissoftware/go-solr-backup:dev-latest
