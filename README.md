# Proxier

Proxier is a better approach to expose applications in Kubernetes. It supports load balancing to a set of pods with weights and provides advanced load balancing strategy by nginx, such as least connections, IP hash.

+ supports canary deployment and load balancing by weight
+ provides advanced load balancing strategy
+ be compatible with kubernetes service behavior
+ scales horizontally with pressure by default

```yaml
apiVersion: maegus.com/v1
kind: Proxier
metadata:
  name: example-proxier
spec:
  ports:
    - name: http
      protocol: TCP
      port: 80
  selector:
    app: example
  backends:
    - name: v1
      weight: 90
      selector:
        version: v1
    - name: v2
      weight: 9
      selector:
        version: v2
```

## Architecture overview

![proxier-architecture](./images/proxier-architecture.png)

## Overview

## Installation

```
```

## License

MIT License, see [LICENSE](./LICENSE)
