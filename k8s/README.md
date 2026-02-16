# Kubernetes Deployment (k3s)

This directory contains Kubernetes manifests for deploying the microservices application to k3s.

## Architecture

```
Namespace: microservices
├── Infrastructure
│   ├── postgres-user      # PostgreSQL for User Service
│   ├── postgres-order     # PostgreSQL for Order Service
│   ├── postgres-payment   # PostgreSQL for Payment Service
│   ├── rabbitmq          # Message queue
│   └── redis             # Cache & rate limiting
├── Config
│   ├── microservices-config   # ConfigMap (env vars)
│   └── microservices-secrets  # Secret (sensitive data)
├── Services
│   ├── user-service      # 2 replicas
│   ├── order-service     # 2 replicas
│   ├── payment-service   # 2 replicas
│   └── gateway           # 1 replica, NodePort:30080
├── Networking
│   └── ingress           # Traefik Ingress (load balancer)
└── Storage
    └── emptyDir          # Non-persistent storage
```

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [k3s](https://k3s.io/) installed and running
- kubectl (included with k3s)
- Make

### Install k3s

```bash
curl -sfL https://get.k3s.io | sh -
```

### Setup kubectl

```bash
# Add kubeconfig to your shell
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# Add to .bashrc for persistence
echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> ~/.bashrc
```

## Quick Start

```bash
# Setup (verify k3s and create namespace)
make k3s-setup

# Build images and load into k3s
make k3s-build

# Deploy
make k3s-deploy

# Or one command
make k3s-up
```

## Access Application

The gateway can be accessed in **two ways**:

### Option 1: NodePort (Simplest)

Direct access via NodePort on port **30080**:

```bash
# From the same machine
curl http://localhost:30080/health

# From other machines in network
curl http://<NODE_IP>:30080/health
```

### Option 2: Ingress with Traefik (Recommended)

k3s includes **Traefik** as the default ingress controller with automatic load balancing:

```bash
# 1. Add host entry
sudo echo "127.0.0.1 microservices.local" >> /etc/hosts

# 2. Access via Ingress
curl http://microservices.local/health
```

**Benefits of Ingress:**
- Load balancing across multiple gateway replicas
- SSL/TLS termination (HTTPS)
- Path-based routing
- Host-based routing
- Automatic retries and failover

**Verify Ingress:**
```bash
# Check ingress status
kubectl get ingress -n microservices

# Test via ingress (requires /etc/hosts entry)
curl -H "Host: microservices.local" http://localhost/health
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make k3s-setup` | Verify k3s and create namespace |
| `make k3s-build` | Build and load images into k3s |
| `make k3s-deploy` | Deploy to k3s |
| `make k3s-up` | Build and deploy (one command) |
| `make k3s-status` | Check pod status |
| `make k3s-logs-gateway` | View gateway logs |
| `make k3s-logs-user` | View user service logs |
| `make k3s-logs-order` | View order service logs |
| `make k3s-logs-payment` | View payment service logs |
| `make k3s-restart-gateway` | Restart gateway |
| `make k3s-scale-user REPLICAS=3` | Scale user service |
| `make k3s-clean` | Delete namespace (cleanup) |

## Manual Steps

```bash
# 1. Setup
./scripts/setup-k3s.sh

# 2. Build images
./scripts/build-k3s-images.sh

# 3. Deploy
./scripts/deploy.sh
```

## Configuration

### Secrets

Update `03-secrets.yaml` with your own values:

```bash
# Generate base64
echo -n 'your-secret' | base64

# Apply
kubectl apply -f k8s/03-secrets.yaml
```

### Environment Variables

Modify `02-configmap.yaml` for non-sensitive configuration.

## Troubleshooting

### k3s not running
```bash
sudo systemctl status k3s
sudo systemctl start k3s
```

### kubectl not connected
```bash
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
kubectl get nodes
```

### Images not loading
```bash
# Check if images are in k3s
sudo k3s ctr images list | grep localhost

# Manual import
sudo k3s ctr images import gateway.tar
```

### Pods not starting
```bash
make k3s-status
kubectl describe pod <pod-name> -n microservices
kubectl logs <pod-name> -n microservices
```

## Cleanup

```bash
# Delete all resources
make k3s-clean

# Or manually
kubectl delete namespace microservices
```

## File Structure

```
k8s/
├── 00-namespace.yaml          # Namespace definition
├── 02-configmap.yaml          # Environment variables
├── 03-secrets.yaml            # Sensitive data
├── 10-postgres-user.yaml      # User database
├── 11-postgres-order.yaml     # Order database
├── 12-postgres-payment.yaml   # Payment database
├── 20-rabbitmq.yaml           # Message queue
├── 21-redis.yaml              # Cache
├── 30-user-service.yaml       # User service
├── 31-order-service.yaml      # Order service
├── 32-payment-service.yaml    # Payment service
├── 40-gateway.yaml            # API Gateway (NodePort:30080)
├── 50-ingress.yaml            # Traefik Ingress (load balancer)
└── README.md                  # This file
```

## Notes

- **Storage**: Uses `emptyDir` (data lost on restart)
- **Auto-migration**: Enabled (services run migrations on startup)
- **Health probes**: Configured for all services
- **NodePort**: Gateway exposed on port 30080 (direct access)
- **Ingress**: Traefik ingress for load balancing and SSL
- **Load Balancing**: Automatic across gateway replicas via Ingress
