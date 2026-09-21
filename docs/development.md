# Development

Use Go 1.27.1 or later. The integration test stack uses Kubernetes 1.37.0, Capsule 0.14.6, and Argo CD 3.5.3.

The Go dependencies use Kubernetes 0.36.4 and controller-runtime 0.24.1 to match Argo CD’s kubectl API. Keep the Kubernetes modules aligned when upgrading them. Argo CD’s `gitops-engine` replacement must point to the same release commit as Argo CD.

Getting started locally is pretty easy. You can execute:

```shell
make e2e-build
```

This installs all required operators an installs the operator within a [KinD Cluster](https://kind.sigs.k8s.io/). The required binaries are also downloaded.

If you wish to test against a specific Kubernetes version, you can pass that via variable:

```shell
KUBERNETES_SUPPORTED_VERSION="v1.37.0" make e2e-build
```

When you want to quickly develop, you can scale down the operator within the cluster:

```shell
kubectl scale deploy capsule-argo-addon --replicas=0 -n capsule-argo-addon
```

And then execute the binary:

```shell
go run ./cmd -zap-log-level=10
```

You might need to first export the Kubeconfig for the cluster (If you are using multiple clusters at the same time):

```shell
bin/kind get kubeconfig --name capsule-arg-addon  > /tmp/capsule-argo-addon
export KUBECONFIG="/tmp/capsule-argo-addon"
```

## Testing

When you are done with the development run the following commands.

For Liniting

```shell
make golint
```

For Unit-Testing

```shell
make test
```

For end-to-end testing, use the dedicated test cluster created by `make e2e-build`:

```shell
make e2e-exec
```

To run only the permission-source scenarios against that test cluster:

```shell
make ginkgo
bin/ginkgo -vv --label-filter=rbac-permissions ./e2e
```

These cases exercise Capsule-resolved User and Group owners, role changes and revocation, rule bindings, and deduplication with legacy additional role bindings. They check the resulting tenant policy in Argo CD’s RBAC ConfigMap.

## Helm Chart

When making changes to the Helm-Chart, Update the documentation by running:

```shell
make helm-docs
```

Linting and Testing the chart:

```shell
make helm-lint
make helm-test
```

## Performance

Use [PProf](https://book.kubebuilder.io/reference/pprof-tutorial) for profiling:

```shell
curl -s "http://127.0.0.1:8082/debug/pprof/profile" > ./cpu-profile.out

go tool pprof -http=:8080 ./cpu-profile.out
```
