#!/bin/bash

if [[ $(docker inspect --format="{{ .State.Running }}" dp-table-renderer) == "false" ]]; then
  exit 1;
fi
