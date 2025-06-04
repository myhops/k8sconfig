# k8sconfig

k8sconfig implements a provider that reads information from Kubernetes secrets and config maps. 
It works similar to the [env provider](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider/envprovider).
It requires that the otel collector has sufficient access privileges in the cluster.

This example shows how the provider can read an item from a secret and a config map.

```yaml
receiver:
  otlp:
    grpc:
      endpoint: ${k8sconfig:secret:my-namespace:my-secret:stringData:key1}
    http:
      endpoint: ${k8sconfig:configMap:my-namespace:my-config-map:data:key1}
```



## References

* [OpenTelemetry Collector Github](https://github.com/open-telemetry/opentelemetry-collector)
* [OpenTelemetry Collector Env Provider](https://github.com/open-telemetry/opentelemetry-collector/tree/main/confmap/provider/envprovider)

