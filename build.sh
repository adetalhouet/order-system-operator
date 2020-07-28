#!/bin/bash

set -x

make docker-build
make docker-push

kubectl delete deployments.apps order-system-operator-controller-manager
sleep 5

make deploy