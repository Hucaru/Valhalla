# Kubernetes Setup Guide

This guide covers deploying Valhalla to a Kubernetes cluster using Helm.

## Why Kubernetes?

Kubernetes provides:
- **High availability** - Automatic restarts and health checks
- **Scalability** - Easy to scale channels horizontally
- **Production-ready** - Battle-tested orchestration platform
- **Infrastructure as code** - Declarative configuration

## Prerequisites

- **Kubernetes cluster** - minikube, kind, K3s, or cloud provider (GKE, EKS, AKS)
- **kubectl** - Configured to connect to your cluster
- **Helm 3** - Package manager for Kubernetes
- **Data.nx file** - See [Installation Guide](Installation.md) for conversion
- **Container registry** - Or ability to load images directly (minikube, kind)

## Quick Start

### Step 1: Prepare Your Cluster

#### Option A: Local Development with minikube

```bash
# Install minikube
# See https://minikube.sigs.k8s.io/docs/start/

# Start cluster
minikube start --memory=4096 --cpus=2

# Verify
kubectl get nodes
```

#### Option B: Local Development with kind

```bash
# Install kind
# See https://kind.sigs.k8s.io/docs/user/quick-start/

# Create cluster
kind create cluster --name valhalla

# Verify
kubectl get nodes
```

#### Option C: Cloud Provider

Use your cloud provider's tools:
- **GKE**: `gcloud container clusters create valhalla`
- **EKS**: Use eksctl or AWS console
- **AKS**: `az aks create --name valhalla`

### Step 2: Prepare Data.nx

For Kubernetes, you need to make Data.nx available to pods. You have two options:

#### Option A: ConfigMap (For Small Files < 1MB)

```bash
kubectl create configmap data-nx --from-file=Data.nx=./Data.nx -n valhalla
```

#### Option B: Persistent Volume (Recommended)

Create a PersistentVolume and copy Data.nx to it. Example for local development:

```yaml
# data-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: data-nx-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadOnlyMany
  hostPath:
    path: /data/valhalla
    type: DirectoryOrCreate
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-nx-pvc
  namespace: valhalla
spec:
  accessModes:
    - ReadOnlyMany
  resources:
    requests:
      storage: 1Gi
```

Apply and copy file:
```bash
kubectl apply -f data-pv.yaml

# For minikube
minikube ssh
sudo mkdir -p /data/valhalla
# Then copy Data.nx to /data/valhalla/ using your preferred method

# For kind - mount when creating cluster
kind create cluster --config=kind-config.yaml
```

### Step 3: Build and Load the Image

```bash
# Build image
docker build -t valhalla:latest -f Dockerfile .

# Load into cluster
# For kind:
kind load docker-image valhalla:latest

# For minikube:
minikube image load valhalla:latest

# For cloud providers:
# Tag and push to your container registry
docker tag valhalla:latest gcr.io/your-project/valhalla:latest
docker push gcr.io/your-project/valhalla:latest
```

### Step 4: Deploy with Helm

```bash
# Create namespace
kubectl create namespace valhalla

# Install chart
helm install valhalla ./helm -n valhalla

# Watch pods start
kubectl get pods -n valhalla -w
```

### Step 5: Expose Services

By default, all services use ClusterIP (internal only). To access from outside:

#### Option A: Port Forwarding (Development)

```bash
# Forward login server
kubectl port-forward -n valhalla svc/login-server 8484:8484

# In another terminal, forward channels
kubectl port-forward -n valhalla svc/channel-server-1 8685:8685
kubectl port-forward -n valhalla svc/channel-server-2 8686:8686
```

#### Option B: LoadBalancer (Cloud)

Edit `helm/values.yaml`:
```yaml
services:
  login:
    type: LoadBalancer
  channels:
    type: LoadBalancer
```

Upgrade:
```bash
helm upgrade valhalla ./helm -n valhalla
```

Get external IPs:
```bash
kubectl get svc -n valhalla
```

#### Option C: Ingress-Nginx (Recommended for Production)

See [Exposing via Ingress](#exposing-services-with-ingress-nginx) below.

## Helm Chart Configuration

### values.yaml

The Helm chart can be customized via `helm/values.yaml`:

```yaml
# Image configuration
image:
  repository: valhalla
  tag: latest
  pullPolicy: IfNotPresent

# Replica counts
replicaCount:
  login: 1
  world: 1
  cashshop: 1
  channels: 2

# Database configuration
database:
  address: "db"
  port: "3306"
  user: "root"
  password: "password"
  database: "maplestory"

# World settings
world:
  message: "Welcome to Valhalla!"
  ribbon: 2
  expRate: 1.0
  dropRate: 1.0
  mesosRate: 1.0

# Channel settings
channel:
  maxPop: 250
  clientConnectionAddress: "127.0.0.1"
```

### Installing with Custom Values

```bash
# Create custom values file
cat > my-values.yaml <<EOF
world:
  expRate: 2.0
  dropRate: 1.5
  mesosRate: 1.5
  
channel:
  maxPop: 500
EOF

# Install with custom values
helm install valhalla ./helm -n valhalla -f my-values.yaml
```

### Upgrading Configuration

```bash
# Edit values.yaml or create new values file
vim helm/values.yaml

# Upgrade deployment
helm upgrade valhalla ./helm -n valhalla

# Rollback if needed
helm rollback valhalla -n valhalla
```

## Service Discovery

In Kubernetes, services use DNS names instead of IP addresses:

| Service | Docker Compose | Kubernetes |
|---------|---------------|------------|
| Login Server | `login_server` | `login-server` |
| World Server | `world_server` | `world-server` |
| Database | `db` | `db` |
| Channel 1 | `channel_server_1` | `channel-server-1` |

The Helm chart automatically adjusts configurations to use hyphens for K8s service names.

## Exposing Services with Ingress-Nginx

Ingress-Nginx allows you to expose TCP services (required for MapleStory):

### Step 1: Install Ingress-Nginx

```bash
# Add Helm repo
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Install ingress-nginx with TCP service support
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --create-namespace \
  --namespace ingress-nginx \
  -f ingress-values.yaml
```

### Step 2: Create ingress-values.yaml

```yaml
tcp:
  8484: valhalla/login-server:8484
  8600: valhalla/cashshop-server:8600
  8685: valhalla/channel-server-1:8685
  8686: valhalla/channel-server-2:8686
  # Add more channels as needed:
  # 8687: valhalla/channel-server-3:8687
  # 8688: valhalla/channel-server-4:8688
```

**Important**: Each channel needs its own port mapping. Port numbers decrease by 1 for each additional channel.

### Step 3: Get External IP

```bash
kubectl get svc -n ingress-nginx ingress-nginx-controller

# Look for EXTERNAL-IP
# On cloud providers, this will be a public IP or hostname
# On minikube: minikube tunnel (in separate terminal)
# On kind: Use port mappings defined in cluster config
```

### Step 4: Update Valhalla Configuration

Update `helm/values.yaml` with the external IP:

```yaml
channel:
  clientConnectionAddress: "<loadbalancer-ip>"
  
cashshop:
  clientConnectionAddress: "<loadbalancer-ip>"
```

Upgrade Helm deployment:
```bash
helm upgrade valhalla ./helm -n valhalla -f helm/values.yaml
```

### Step 5: Update MapleStory Client

Configure your client to connect to the ingress controller's external IP.

## Scaling Channels

### Add More Channels

Edit `helm/values.yaml`:
```yaml
replicaCount:
  channels: 5  # Increase from 2 to 5
```

Update ingress-values.yaml to include new channel ports:
```yaml
tcp:
  8484: valhalla/login-server:8484
  8600: valhalla/cashshop-server:8600
  8685: valhalla/channel-server-1:8685
  8686: valhalla/channel-server-2:8686
  8687: valhalla/channel-server-3:8687
  8688: valhalla/channel-server-4:8688
  8689: valhalla/channel-server-5:8689
```

Upgrade both:
```bash
# Upgrade ingress
helm upgrade ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx \
  -f ingress-values.yaml

# Upgrade valhalla
helm upgrade valhalla ./helm -n valhalla
```

## Database

### Using External MySQL

For production, use a managed database service:

```yaml
# values.yaml
database:
  address: "mysql.example.com"
  port: "3306"
  user: "valhalla"
  password: "securePassword"
  database: "maplestory"
```

### Using In-Cluster MySQL

The Helm chart can deploy MySQL within the cluster (not recommended for production):

```yaml
mysql:
  enabled: true
  persistence:
    enabled: true
    size: 10Gi
```

## Monitoring

### Prometheus

Install Prometheus to scrape Valhalla metrics:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
```

Configure ServiceMonitor:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: valhalla
  namespace: valhalla
spec:
  selector:
    matchLabels:
      app: valhalla
  endpoints:
    - port: metrics
      path: /metrics
```

### Grafana

Access Grafana (installed with Prometheus):
```bash
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

Default credentials: admin/prom-operator

## Managing the Deployment

### View Pods

```bash
kubectl get pods -n valhalla
```

### View Logs

```bash
# Specific pod
kubectl logs -n valhalla login-server-xyz123

# Follow logs
kubectl logs -n valhalla -f login-server-xyz123

# All pods with label
kubectl logs -n valhalla -l app=channel-server
```

### Execute Commands in Pod

```bash
kubectl exec -it -n valhalla channel-server-1-xyz123 -- sh
```

### Restart Deployment

```bash
kubectl rollout restart deployment/login-server -n valhalla
```

### Scale Manually

```bash
kubectl scale deployment/channel-server --replicas=3 -n valhalla
```

## Troubleshooting

### Pods Not Starting

**Check pod status**:
```bash
kubectl describe pod -n valhalla <pod-name>
```

**Common issues**:
- Image pull error: Check image name and pull policy
- Missing Data.nx: Verify ConfigMap or PV is correctly mounted
- Database connection: Check database service and credentials

### CrashLoopBackOff

**Check logs**:
```bash
kubectl logs -n valhalla <pod-name> --previous
```

**Common causes**:
- Missing environment variables
- Database not ready
- Incorrect configuration

### Service Not Reachable

**Check service**:
```bash
kubectl get svc -n valhalla
kubectl describe svc -n valhalla login-server
```

**Test connectivity**:
```bash
# From inside cluster
kubectl run -it --rm debug --image=alpine --restart=Never -n valhalla -- sh
apk add curl netcat-openbsd
nc -zv login-server 8484
```

### ConfigMap/Secret Changes Not Reflected

Pods don't automatically restart when ConfigMaps/Secrets change:

```bash
# Force restart
kubectl rollout restart deployment/login-server -n valhalla
```

## Security Best Practices

1. **Use Secrets for sensitive data**:
   ```bash
   kubectl create secret generic db-credentials \
     --from-literal=password=securePassword \
     -n valhalla
   ```

2. **Set resource limits**:
   ```yaml
   resources:
     limits:
       memory: "512Mi"
       cpu: "500m"
     requests:
       memory: "256Mi"
       cpu: "250m"
   ```

3. **Use RBAC** for access control
4. **Enable Network Policies** to restrict traffic
5. **Run as non-root user** where possible
6. **Keep images updated** regularly

## Backup and Recovery

### Backup Database

```bash
# If using in-cluster MySQL
kubectl exec -n valhalla db-0 -- mysqldump -u root -ppassword maplestory > backup.sql

# If using PVC
kubectl exec -n valhalla db-0 -- mysqldump -u root -ppassword maplestory | gzip > backup.sql.gz
```

### Restore Database

```bash
kubectl exec -i -n valhalla db-0 -- mysql -u root -ppassword maplestory < backup.sql
```

## Production Checklist

- [ ] Use managed database service
- [ ] Set up SSL/TLS certificates
- [ ] Configure resource requests and limits
- [ ] Set up monitoring and alerting
- [ ] Configure automatic backups
- [ ] Use Secrets for sensitive data
- [ ] Set up logging aggregation
- [ ] Configure pod disruption budgets
- [ ] Test disaster recovery procedures
- [ ] Set up autoscaling (HPA) if needed
- [ ] Configure network policies
- [ ] Use separate namespaces for different environments

## Next Steps

- Configure server settings: [Configuration.md](Configuration.md)
- Learn about Docker deployment: [Docker.md](Docker.md)
- Understand local development: [Local.md](Local.md)
- Build from source: [Building.md](Building.md)

## Useful Commands Reference

```bash
# Deploy
helm install valhalla ./helm -n valhalla

# Upgrade
helm upgrade valhalla ./helm -n valhalla

# Rollback
helm rollback valhalla -n valhalla

# Uninstall
helm uninstall valhalla -n valhalla

# View values
helm get values valhalla -n valhalla

# Check status
helm status valhalla -n valhalla

# View pods
kubectl get pods -n valhalla

# View services
kubectl get svc -n valhalla

# View logs
kubectl logs -f -n valhalla <pod-name>

# Port forward
kubectl port-forward -n valhalla svc/login-server 8484:8484

# Execute in pod
kubectl exec -it -n valhalla <pod-name> -- sh
```
