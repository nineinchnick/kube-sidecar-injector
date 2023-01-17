# kube-sidecar-injector

This repo contains a Kubernetes [MutatingAdmissionWebhook](https://kubernetes.io/docs/admin/admission-controllers/#mutatingadmissionwebhook-beta-in-19) that injects a sidecar container into pod prior to persistence of the object.

## Prerequisites

- [git](https://git-scm.com/downloads)
- [go](https://golang.org/dl/) version v1.17+
- [docker](https://docs.docker.com/install/) version 19.03+
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) version v1.19+
- Access to a Kubernetes v1.19+ cluster with the `admissionregistration.k8s.io/v1` API enabled. Verify that by the following command:

```bash
kubectl api-versions | grep admissionregistration.k8s.io
```

The result should be:

```
admissionregistration.k8s.io/v1
admissionregistration.k8s.io/v1beta1
```

> Note: In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.

## Build

Build and push docker image

```bash
export DOCKER_USER=nineinchnick
./build
```

## Deploy

Modify the files in the `deployment` directory as you see fit, especially the config map. Then apply them using a command like this:

```bash
kubectl apply -k deployment
```

## Troubleshooting

Sometimes you may find that pod is injected with sidecar container as expected, check the following items:

1. The sidecar-injector pod is in running state and no error logs.
2. The namespace in which application pod is deployed has the correct labels(`sidecar-injector=enabled`) as configured in `mutatingwebhookconfiguration`.
3. Check if the application pod has annotation `sidecar-injector-webhook.was.net.pl/inject:"yes"`.
