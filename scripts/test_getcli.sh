#!/bin/bash

set -xe

cd $(dirname $0)
f="$PWD/getcli.sh"

docker run -v "$f:/cli" --rm centos sh -c "sh /cli && banzai --version"

docker run -v "$f:/cli" --rm centos sh -c "sh </cli && banzai --version"
docker run -v "$f:/cli" --rm ubuntu sh -c "apt-get update; env TERM=dumb apt-get -y install --no-install-recommends ca-certificates wget; sh </cli && banzai --version"
docker run -v "$f:/cli" --rm ubuntu:16.04 sh -c "apt-get update; env TERM=dumb apt-get -y install --no-install-recommends ca-certificates wget; sh </cli && banzai --version"
docker run -v "$f:/cli" --rm debian sh -c "apt-get update; env TERM=dumb apt-get -y install --no-install-recommends ca-certificates curl; sh </cli && banzai --version"
docker run -v "$f:/cli" --rm golang sh -c "sh </cli && banzai --version"
docker run -v "$f:/cli" --rm fedora sh -c "sh </cli && banzai --version"

docker run -v "$f:/cli" --rm centos sh -c "bash </cli && banzai --version"
docker run -v "$f:/cli" --rm ubuntu sh -c "apt-get update; env TERM=dumb apt-get -y install --no-install-recommends ca-certificates curl; bash </cli && banzai --version"

! docker run -v "$f:/cli" --rm ubuntu sh -c "bash </cli && banzai --version"

docker run -v "$f:/cli" --rm ubuntu sh -c "apt-get update; env TERM=dumb apt-get -y install --no-install-recommends ca-certificates curl; rm /etc/os-release; bash </cli && banzai --version"

echo OK
