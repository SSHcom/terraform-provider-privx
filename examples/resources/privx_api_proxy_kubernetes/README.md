# PrivX API Proxy Kubernetes Example


Prerequisites
* PrivX environment reachable
* PrivX Terraform provider (dev override or installed)
* A Kubernetes API bearer token that PrivX can use


### 🔐 How to verify Certificates

| File                                      | Subject | Issuer |
|:------------------------------------------| :--- | :--- |
| **EKS Cluster CA** (`eks-cluster-ca.pem`) | `CN=kubernetes` | `CN=kubernetes` |
| **PrivX API Proxy** (`privx-ca.pem`)      | `CN=PrivX Api-Proxy CA` | `CN=PrivX Api-Proxy CA` |

#### 🔍 Command Details

**EKS Cluster CA Check**
```bash
openssl x509 -in eks-cluster-ca.pem -noout -subject -issuer
```
> **Subject:** `CN=kubernetes`  
> **Issuer:** `CN=kubernetes`

**PrivX Proxy CA Check**
```bash
openssl x509 -in privx-ca.pem -noout -subject -issuer
```
> **Subject:** `CN=PrivX Api-Proxy CA`  
> **Issuer:** `CN=PrivX Api-Proxy CA`


### 🔐 How to verify secret tokens are valid

**EKS Cluster token check**
```bash
export TF_VAR_kubernetes_api_token=TOKEN
curl --cacert eks-cluster-ca.pem -H "Authorization: Bearer $TF_VAR_kubernetes_api_token"   https://my-eks.amazonaws.com/version
```


**Privx token check**
```bash
export PRIVX_PROXY_CA_PEM=privx-ca.pem
export PRIVX_PROXY_URL=http://my-privx.compute.amazonaws.com:20080
export PRIVX_KUBECTL_TOKEN=""
curl \
--cacert $PRIVX_PROXY_CA_PEM \
-H "Authorization: Bearer $PRIVX_KUBECTL_TOKEN" \
"$PRIVX_PROXY_URL/version"
```



# Operating The "Kubeconfig" Workflow

Step 1: Run terraform apply.

Step 2: Use the outputs (Proxy URL, CA, and Token).

Show readable privx_kubectl_token:
```terraform output -raw privx_kubectl_token```

Step 3: Format the kubeconfig.
```
kubectl --kubeconfig kubeconfig.yaml get all -n privx
```

Example of the kubeconfig.yaml file:
```
apiVersion: v1
clusters:
- cluster:
  server: https://my-eks.amazonaws.com
  proxy-url: http://my-privx.compute.amazonaws.com:20080
  certificate-authority-data: LS0tLS1CRlGSUNBVEUtLS0tLQo=
  name: arn:aws:eks:eu-north-1:606364382:cluster/privx-kube-api
  contexts:
- context:
  cluster: arn:aws:eks:eu-north-1:606364382:cluster/privx-kube-api
  user: arn:aws:eks:eu-north-1:606364382:cluster/privx-kube-api
  name: arn:aws:eks:eu-north-1:606364382:cluster/privx-kube-api
  current-context: arn:aws:eks:eu-north-1:606369954382:cluster/privx-kube-api
  kind: Config
  preferences: {}
  users:
   - name: arn:aws:eks:eu-north-1:606364382:cluster/privx-kube-api
     user:
     token: <SET HERE TOKEN TAKEN FROM THE TF STATE or FROM PRIVX UI>
```
