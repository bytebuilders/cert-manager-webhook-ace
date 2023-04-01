# Cert-Manager ACME DNS01 Webhook Solver for AppsCode DNS Proxy for Cloudfare

This solver can be used when you want to use [cert-manager](https://github.com/cert-manager/cert-manager) with AppsCode DNS Proxy for Cloudfare. 

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Webhook

#### Using public helm chart

```bash
helm repo add appscode https://charts.appscode.com/stable/
helm repo update
helm upgrade -i cert-manager-webhook-ace okteto/cert-manager-webhook-ace \
  --namespace cert-manager --create-namespace
```

**Note**: The kubernetes resources used to install the Webhook should be deployed within the same namespace as the cert-manager.

To uninstall the webhook run
```bash
helm uninstall --namespace cert-manager cert-manager-webhook-ace
```

## Usage

### Credentials

In order to access the DNS Proxy, the webhook needs an [API token] from ByteBuilders.

```
kubectl create secret generic ace-secret --from-literal=api-token=<YOUR_TOKEN>
```

### Create Issuer

Create a `ClusterIssuer` or `Issuer` resource as following:

#### Cluster-wide Issuer
```
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    
    # Email address used for ACME registration
    email: mail@example.com # REPLACE THIS WITH YOUR EMAIL
    
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
    - dns01:
        webhook:
          solverName: "ace"
          groupName: webhook.dns.appscode.com
          config:
            baseURL: "https://dns.byte.builders"
            apiTokenSecretRef:
              name: ace-secret
              key: api-token
```

By default, the API token used will be obtained from the secret in the same namespace as the webhook.

### Create a certificate

Create your certificate resource as follows:

```
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-cert
  namespace: cert-manager
spec:
  commonName: example.com
  dnsNames:
  - example.com # REPLACE THIS WITH YOUR DOMAIN
  issuerRef:
   name: letsencrypt-staging
   kind: ClusterIssuer
  secretName: example-cert
```

# Acknowledgement

This project was forked from [okteto/cert-manager-webhook-civo](https://github.com/okteto/cert-manager-webhook-civo). The cloudflare package was copied from [cert-manager project](https://github.com/cert-manager/cert-manager/tree/master/pkg/issuer/acme/dns/cloudflare). 
