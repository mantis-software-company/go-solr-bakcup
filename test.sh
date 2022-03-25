#!/bin/sh
TARGET_HOST=$(printenv CONTAINER_HOST|sed 's-tcp://--g'|cut -f1 -d:)
rm -fr tmp/*

find . -maxdepth 1 -name "*.test" -delete
REV=$(git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD)
NOW=$(date +'%Y-%m-%d_%T')
go mod tidy
go mod vendor
for pkg in `go list ./...`;
do
  go test -cover -c $pkg
done

if [ "x$1" == "xrun" ]; then
  for tf in $(find . -type f -name "*.test");
  do
    ./$tf -test.coverprofile tmp/cover.out -ginkgo.v -test.v -v ${VERBOSE:-5} ||exit 1
    go tool cover -html tmp/cover.out -o tmp/cover.html
  done
fi
