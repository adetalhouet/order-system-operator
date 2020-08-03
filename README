
## TODO
- istio inject
- apigw

## Prerequisites

### Postgres & NATS
The order system rely on a postgres database as well as a NATS message bus.
If you already have an instance deployed, available to consume, then skip this step.
Else, you can deploy them as follow:
```
helm install order-system-db -n default \
  --set postgresqlUsername=order-system,postgresqlPassword=Password123,postgresqlDatabase=order-system \
  bitnami/postgresql

helm install order-system-nats -n default \
  --set auth.enabled=true,auth.user=order-system,auth.password=Password123 \
    bitnami/nats
```

NATS & DB credentials must be provided through a secret, using the following keys:
- username
- password

Example, in the case they are sharing the same credentials; else, create two secrets:

```
kubectl create secret generic order-system-crendentials --from-literal=username=order-system --from-literal=password=Password123
```


### Support for HPA

1. Instal the hpa-operator from Banzai

```
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com
helm repo udpate
helm install banzaicloud-stable/hpa-operator
```

## Install
...

## Apply CRD
`kubectl apply -f example/order-system.yaml`
