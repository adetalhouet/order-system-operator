module github.com/adetalhouet/order-system-operator

go 1.13

require (
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/redhat-cop/operator-utils v0.3.3
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2
