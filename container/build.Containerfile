FROM golang:1.16-alpine as builder
RUN apk add make git build-base protoc --no-cache \
    && go get google.golang.org/protobuf/cmd/protoc-gen-go \
    && wget -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_64.zip \
    && unzip protoc.zip && rm protoc.zip && rm bin/protoc && mv /usr/bin/protoc bin/protoc

COPY . /source
WORKDIR /source
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV LDFLAGS="-w -s"
RUN make build

FROM scratch
COPY --from=builder /source/bin/solr-backup /bin/solr-backup
ENTRYPOINT ["/bin/solr-backup"]
