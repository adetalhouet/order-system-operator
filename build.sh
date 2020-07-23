#!/bin/bash

set -x

make docker-build
make docker-push
make deploy


kubectl get pods -n order-system-operator-system -o=jsonpath='{.items[0].metadata.name}' | xargs kubectl delete pods -n order-system-operator-system
kubectl get pods -n order-system-operator-system -o=jsonpath='{.items[0].metadata.name}' | xargs kubectl logs -f -n order-system-operator-system -c manager