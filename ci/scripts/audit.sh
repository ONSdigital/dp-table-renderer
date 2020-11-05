---
platform: linux

image_resource:
  type: docker-image
  source:
    repository: onsdigital/dp-concourse-tools-nancy
    tag: latest

inputs:
  - name: dp-table-renderer
    path: dp-table-renderer

run:
  path: dp-table-renderer/ci/scripts/audit.sh