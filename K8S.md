# Kubernetes Deployment Guide - k3d Local Cluster

Panduan lengkap deploy microservices ke Kubernetes lokal menggunakan k3d.

## Prerequisites

### Hardware Requirements
- **CPU**: 4 cores minimum, 6+ cores recommended
- **RAM**: 8GB minimum, 12GB+ recommended
- **Disk**: 15GB free space (SSD recommended)
- **Internet**: 1.5-2 GB untuk download pertama kali

### Software Requirements
- Docker Desktop/Engine
- kubectl
- k3d

### Install Prerequisites

```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install k3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Verify installations
kubectl version --client
k3d version
```

## Quick Start

```bash
# 1. Setup cluster dan deploy semua
make k8s-setup
make k8s-build
make k8s-deploy

# 2. Verifikasi
curl http://localhost/health
```

## Detailed Deployment

### Step 1: Setup k3d Cluster

```bash
./scripts/setup-k3d.sh
```

Ini akan:
- Membuat k3d cluster dengan 1 server dan 2 agents
- Setup local registry (microservices-registry:5000)
- Port forward: 80 (HTTP) dan 4000 (gateway direct)
- Buat namespace `microservices`

### Step 2: Build Docker Images

```bash
./scripts/build-images.sh
```

Build dan push 4 images ke local registry:
- gateway:latest
- user-service:latest
- order-service:latest
- payment-service:latest

**Internet usage: 0 MB** (karena local registry)

### Step 3: Deploy Infrastructure

```bash
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/01-storage.yaml
kubectl apply -f k8s/02-configmap.yaml
kubectl apply -f k8s/03-secrets.yaml
```

Deploy:
- PostgreSQL (3 instances: user, order, payment)
- RabbitMQ (message queue)
- Redis (cache & rate limiting)

### Step 4: Deploy Applications

```bash
kubectl apply -f k8s/10-postgres-user.yaml
kubectl apply -f k8s/11-postgres-order.yaml
kubectl apply -f k8s/12-postgres-payment.yaml
kubectl apply -f k8s/20-rabbitmq.yaml
kubectl apply -f k8s/21-redis.yaml

# Wait for infrastructure
kubectl wait --for=condition=ready pod -l app=postgres-user --timeout=120s
kubectl wait --for=condition=ready pod -l app=postgres-order --timeout=120s
kubectl wait --for=condition=ready pod -l app=postgres-payment --timeout=120s
kubectl wait --for=condition=ready pod -l app=rabbitmq --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis --timeout=120s

# Deploy services
kubectl apply -f k8s/30-user-service.yaml
kubectl apply -f k8s/31-order-service.yaml
kubectl apply -f k8s/32-payment-service.yaml
kubectl apply -f k8s/40-gateway.yaml

# Wait for services
kubectl wait --for=condition=ready pod -l app=user-service --timeout=120s
kubectl wait --for=condition=ready pod -l app=order-service --timeout=120s
kubectl wait --for=condition=ready pod -l app=payment-service --timeout=120s
kubectl wait --for=condition=ready pod -l app=gateway --timeout=120s

# Deploy ingress
kubectl apply -f k8s/50-ingress.yaml

# Run migrations
kubectl apply -f k8s/60-migrations.yaml
```

Atau pakai script otomatis:
```bash
./scripts/deploy.sh
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        k3d Cluster                          │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Ingress Controller (LoadBalancer)                    │  │
│  │  Port: 80 → Gateway:4000                              │  │
│  └──────────────────┬────────────────────────────────────┘  │
│                     │                                       │
│  ┌──────────────────▼────────────────────────────────────┐  │
│  │  Namespace: microservices                              │  │
│  │                                                        │  │
│  │  ┌──────────────┐    ┌──────────────┐                 │  │
│  │  │   Gateway    │    │   Ingress    │                 │  │
│  │  │   :4000      │    │   :80        │                 │  │
│  │  └──────┬───────┘    └──────────────┘                 │  │
│  │         │                                              │  │
│  │  ┌──────┴──────┬──────────────┬──────────────┐         │  │
│  │  │             │              │              │         │  │
│  │  ▼             ▼              ▼              ▼         │  │
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐       │  │
│  │  │ User   │  │ Order  │  │ Payment│  │Migrate │       │  │
│  │  │Service │  │Service │  │Service │  │  Job   │       │  │
│  │  │:8081   │  │:8082   │  │:8083   │  │        │       │  │
│  │  └────┬───┘  └────┬───┘  └────┬───┘  └────────┘       │  │
│  │       │           │           │                       │  │
│  │       └───────────┴───────────┘                       │  │
│  │                   │                                   │  │
│  │  ┌────────────────┼────────────────┐                  │  │
│  │  │                │                │                  │  │
│  │  ▼                ▼                ▼                  │  │
│  │  ┌────────┐  ┌────────┐  ┌────────┐                  │  │
│  │  │Postgres│  │RabbitMQ│  │ Redis  │                  │  │
│  │  │-user   │  │:5672   │  │:6379   │                  │  │
│  │  │-order  │  │:15672  │  │        │                  │  │
│  │  │-payment│  │        │  │        │                  │  │
│  │  └────────┘  └────────┘  └────────┘                  │  │
│  │                                                        │  │
│  └────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Access Points

Setelah deploy berhasil:

| Service | URL | Keterangan |
|---------|-----|------------|
| GraphQL Playground | http://localhost | Via Ingress |
| Gateway Direct | http://localhost:4000 | Port-forward |
| RabbitMQ Management | http://localhost:15672 | guest/guest |
| User Service | http://user-service.microservices.svc.cluster.local:8081 | Internal |
| Order Service | http://order-service.microservices.svc.cluster.local:8082 | Internal |
| Payment Service | http://payment-service.microservices.svc.cluster.local:8083 | Internal |

## Makefile Commands

```bash
# Setup & Deployment
make k8s-setup          # Setup k3d cluster
make k8s-build          # Build all Docker images
make k8s-deploy         # Deploy all to Kubernetes
make k8s-deploy-infra   # Deploy infrastructure only
make k8s-deploy-apps    # Deploy applications only

# Operations
make k8s-logs           # View gateway logs
make k8s-logs-user      # View user service logs
make k8s-logs-order     # View order service logs
make k8s-logs-payment   # View payment service logs
make k8s-status         # Check all pods status
make k8s-describe       # Describe all pods

# Development
make k8s-rebuild        # Rebuild and redeploy all
make k8s-restart        # Restart all deployments
make k8s-port-forward   # Port forward gateway to localhost:4000

# Cleanup
make k8s-delete         # Delete k3d cluster
make k8s-clean          # Clean all resources
```

## Troubleshooting

### 1. Pods tidak start
```bash
# Check pod status
kubectl get pods -n microservices

# Check events
kubectl get events -n microservices --sort-by='.lastTimestamp'

# Describe pod
kubectl describe pod <pod-name> -n microservices

# Check logs
kubectl logs <pod-name> -n microservices
```

### 2. Image pull error
```bash
# Pastikan image ada di local registry
docker images | grep microservices-registry

# Rebuild jika perlu
make k8s-build

# Check registry
kubectl get pods -n kube-system | grep registry
```

### 3. Database connection error
```bash
# Check if postgres is ready
kubectl get pods -l app=postgres-user -n microservices

# Check logs
kubectl logs -l app=postgres-user -n microservices

# Exec ke postgres
kubectl exec -it postgres-user-0 -n microservices -- psql -U postgres -d user
```

### 4. Gateway tidak bisa diakses
```bash
# Check ingress
kubectl get ingress -n microservices

# Check gateway service
kubectl get svc gateway -n microservices

# Port forward manual
kubectl port-forward svc/gateway 4000:4000 -n microservices

# Test
curl http://localhost:4000/health
```

### 5. Migration gagal
```bash
# Check migration job
kubectl get jobs -n microservices

# Check migration logs
kubectl logs job/migration-job -n microservices

# Re-run migration
kubectl delete job migration-job -n microservices
kubectl apply -f k8s/60-migrations.yaml
```

## Storage

### Persistent Volumes

| Storage | Size | Purpose |
|---------|------|---------|
| postgres-user-data | 5Gi | User database |
| postgres-order-data | 5Gi | Order database |
| postgres-payment-data | 5Gi | Payment database |
| rabbitmq-data | 2Gi | Message queue persistence |
| redis-data | 1Gi | Cache persistence |

### Data Backup (Opsional)

```bash
# Backup postgres
kubectl exec -it postgres-user-0 -n microservices -- pg_dump -U postgres user > backup-user.sql

# Restore
kubectl exec -i postgres-user-0 -n microservices -- psql -U postgres user < backup-user.sql
```

## Monitoring

```bash
# Watch pods
watch kubectl get pods -n microservices

# Resource usage
kubectl top nodes
kubectl top pods -n microservices

# Check resource limits
kubectl describe node
```

## Cleanup

```bash
# Delete semua resources tapi keep cluster
kubectl delete -f k8s/

# Delete cluster sepenuhnya
k3d cluster delete microservices

# Delete registry
docker rm -f k3d-microservices-registry

# Clean volumes
docker volume prune
```

## Tips

1. **Development Mode**: Gunakan `make k8s-rebuild` untuk rebuild cepat setelah code changes
2. **Logs**: Gunakan `stern` untuk log aggregation jika sudah install
3. **Debug**: Port-forward service tertentu untuk debug langsung
4. **Scale**: Edit replica count di deployment YAML untuk scale up/down

## Resource Limits per Pod

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---------|-------------|-----------|----------------|--------------|
| gateway | 100m | 500m | 128Mi | 256Mi |
| user-service | 100m | 300m | 128Mi | 256Mi |
| order-service | 100m | 300m | 128Mi | 256Mi |
| payment-service | 100m | 300m | 128Mi | 256Mi |
| postgres-* | 100m | 500m | 256Mi | 512Mi |
| rabbitmq | 100m | 300m | 256Mi | 512Mi |
| redis | 50m | 200m | 64Mi | 128Mi |

**Total estimated resources:**
- CPU: 1-2 cores
- Memory: 2-3 GB

## Next Steps

Setelah deployment berhasil:
1. Test GraphQL playground di http://localhost
2. Buat user test
3. Create order dan payment
4. Monitor logs dengan `make k8s-logs`

Selamat mencoba! 🚀
