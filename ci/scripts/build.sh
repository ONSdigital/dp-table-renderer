#!/bin/bash -eux

pushd dp-table-renderer
  make build
  cp build/dp-table-renderer Dockerfile.concourse ../build
popd
