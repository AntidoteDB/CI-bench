#! /bin/bash

cd $(dirname $0)
ARGS="$*" docker-compose  up
docker-compose  down -v