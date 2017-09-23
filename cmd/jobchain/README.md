# jobchain

Exits with 0 only when the job completes successfully.

```
jobchain -name JOBNAME [-namespace NAMESPACE]
```

# Kubernetes

Use `jobchain` to simulate sequential jobs by setting it as the `initcontainer`
for the job's pod.

```yaml
initContainers:
  - name: jobchain
    image: carolynvs/jobchain
    args:
    - "--namespace"
    - "NAMESPACE"
    - "--name"
    - "JOBNAME"
```

