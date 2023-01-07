# Cert-Manager ACME DNS01 Webhook Solver for AppsCode DNS Proxy for Cloudflare

[![Go Report Card](https://goreportcard.com/badge/github.com/bytebuilders/cert-manager-webhook-appscode)](https://goreportcard.com/report/github.com/bytebuilders/cert-manager-webhook-appscode)
[![Releases](https://img.shields.io/github/v/release/bytebuilders/cert-manager-webhook-appscode?include_prereleases)](https://github.com/bytebuilders/cert-manager-webhook-appscode/releases)
[![LICENSE](https://img.shields.io/github/license/bytebuilders/cert-manager-webhook-appscode)](https://github.com/slicen/cert-manager-webhook-appscode/blob/master/LICENSE)
[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/okteto)](https://artifacthub.io/packages/search?repo=okteto)

This solver can be used when you want to use  [cert-manager](https://github.com/cert-manager/cert-manager) with AppsCode DNS Proxy for Cloudflare. 

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Webhook

#### Using public helm chart

```bash
helm repo add okteto https://charts.okteto.com
helm repo update
helm install --namespace cert-manager cert-manager-webhook-appscode bytebuilders/cert-manager-webhook-appscode
```

#### From local checkout

```bash
helm install --namespace cert-manager cert-manager-webhook-appscode chart/cert-manager-webhook-appscode
```
**Note**: The kubernetes resources used to install the Webhook should be deployed within the same namespace as the cert-manager.

To uninstall the webhook run
```bash
helm uninstall --namespace cert-manager cert-manager-webhook-appscode
```

## Usage

### Credentials
In order to access the CIVO API, the webhook needs an [API token](https://www.civo.com/account/security).

```
kubectl create secret generic ace-secret --from-literal=api-key=<YOUR_CIVO_TOKEN>
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
          solverName: "appscode"
          groupName: dns-proxy.appscode.com
          config:
            secretName: ace-secret
```

By default, the CIVO API token used will be obtained from the secret in the same namespace as the webhook.

#### Per Namespace API Tokens

If you would prefer to use separate API tokens for each namespace (e.g. in a multi-tenant environment):

```
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-staging
  namespace: default
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
          solverName: "appscode"
          groupName: dns-proxy.appscode.com
          config:
            secretName: ace-secret
```

By default, the webhook doesn't have permissions to read secrets on all namespaces. To enable this, you'll need to provide your own service account.

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

Forked from https://github.com/okteto/cert-manager-webhook-civo
