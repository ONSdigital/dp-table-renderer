#!/bin/bash

AWS_REGION=
CONFIG_BUCKET=
ECR_REPOSITORY_URI=
GIT_COMMIT=

INSTANCE=$(curl -s http://instance-data/latest/meta-data/instance-id)
CONFIG=$(aws --region $AWS_REGION ec2 describe-tags --filters "Name=resource-id,Values=$INSTANCE" "Name=key,Values=Configuration" --output text | awk '{print $5}')

(aws s3 cp s3://$CONFIG_BUCKET/frontend-router/$CONFIG.asc . && gpg --decrypt $CONFIG.asc > $CONFIG) || exit $?

source $CONFIG && docker run -d                    \
  --env=BIND_ADDR=$BIND_ADDR                       \
  --env=CORS_ALLOWED_ORIGINS=$CORS_ALLOWED_ORIGINS \
  --env=SHUTDOWN_TIMEOUT=$SHUTDOWN_TIMEOUT         \
  --name=dp-table-renderer                         \
  --net=$DOCKER_NETWORK                            \
  --restart=always                                 \
  $ECR_REPOSITORY_URI/dp-table-renderer:$GIT_COMMIT
