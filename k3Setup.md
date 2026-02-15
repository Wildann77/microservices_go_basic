Rencana Deployment ke k3d (Local Kubernetes)

Berikut rencana lengkap deploy project ini ke Kubernetes lokal:

---

### **Phase 1: Persiapan**


**2. Setup k3d Cluster**
```bash
# Create cluster dengan:
# - 1 server (control plane)
# - 2 agents (worker nodes)
# - Built-in registry untuk local images
# - Port mapping untuk gateway (80 & 4000)
# - Sufficient resources untuk semua services

k3d cluster create microservices \
  --servers 1 \
  --agents 2 \
  --port "80:80@loadbalancer" \
  --port "4000:4000@loadbalancer" \
  --registry-create microservices-registry:5000 \
  --volume $(pwd):/src@all \
  --wait
```

---

### **Phase 2: Struktur Namespace & Resources**

```
Namespace: microservices
├── Infrastructure
│   ├── postgres-user (StatefulSet + PVC)
│   ├── postgres-order (StatefulSet + PVC)
│   ├── postgres-payment (StatefulSet + PVC)
│   ├── rabbitmq (StatefulSet)
│   └── redis (Deployment)
├── Config
│   ├── configmap-apps (env vars)
│   └── secret-apps (passwords, keys)
├── Services
│   ├── user-service (Deployment + Service)
│   ├── order-service (Deployment + Service)
│   ├── payment-service (Deployment + Service)
│   └── gateway (Deployment + Service)
└── Networking
    └── ingress (Ingress resource)
```

---

### **Phase 3: Infrastructure Deployment**

**Database PostgreSQL (3 instances):**
- Pakai **StatefulSet** untuk data persistence
- **PersistentVolumeClaim** 5GB per database
- Service type: ClusterIP (internal only)
- Init container untuk run migrations
- Health checks dengan pg_isready

**Message Queue RabbitMQ:**
- StatefulSet dengan 1 replica (dev mode)
- Management UI: port 15672
- AMQP: port 5672
- PersistentVolume untuk message durability

**Cache Redis:**
- Deployment dengan 1 replica (dev mode)
- PersistentVolume untuk AOF
- Service: ClusterIP

---

### **Phase 4: Application Deployment**

**Build & Push Images:**
```bash
# Build semua images
docker build -t microservices-registry:5000/gateway:latest ./gateway
docker build -t microservices-registry:5000/user-service:latest ./services/user
docker build -t microservices-registry:5000/order-service:latest ./services/order
docker build -t microservices-registry:5000/payment-service:latest ./services/payment

# Push ke local registry
docker push microservices-registry:5000/gateway:latest
docker push microservices-registry:5000/user-service:latest
docker push microservices-registry:5000/order-service:latest
docker push microservices-registry:5000/payment-service:latest
```

**Deployment Specs:**

| Service | Replicas | Port | Resources |
|---------|----------|------|-----------|
| gateway | 1 | 4000 | 256Mi/512Mi |
| user-service | 2 | 8081 | 256Mi/512Mi |
| order-service | 2 | 8082 | 256Mi/512Mi |
| payment-service | 2 | 8083 | 256Mi/512Mi |

**ConfigMap untuk environment variables:**
- Database connection strings
- RabbitMQ URL
- Redis URL
- Service URLs (internal communication)
- JWT config
- Rate limiting config
- Cache TTL

**Secret untuk sensitive data:**
- DB passwords
- JWT secret key
- Stripe API keys
- RabbitMQ credentials

---

### **Phase 5: Networking**

**Services:**
- Semua services pakai **ClusterIP** (internal only)
- Gateway expose via **Ingress**

**Ingress Setup:**
```yaml
# Host-based routing
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gateway-ingress
spec:
  rules:
  - host: microservices.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gateway
            port:
              number: 4000
```

**Akses lokal:**
Tambahkan ke `/etc/hosts`:
```
127.0.0.1 microservices.local
```

---

### **Phase 6: Deployment Order**

```bash
# 1. Namespace
kubectl apply -f k8s/00-namespace.yaml

# 2. Storage (PV/PVC)
kubectl apply -f k8s/01-storage.yaml

# 3. Config & Secrets
kubectl apply -f k8s/02-configmap.yaml
kubectl apply -f k8s/03-secrets.yaml

# 4. Infrastructure
kubectl apply -f k8s/10-postgres-user.yaml
kubectl apply -f k8s/11-postgres-order.yaml
kubectl apply -f k8s/12-postgres-payment.yaml
kubectl apply -f k8s/20-rabbitmq.yaml
kubectl apply -f k8s/21-redis.yaml

# Wait for infrastructure ready
kubectl wait --for=condition=ready pod -l app=postgres-user --timeout=120s
kubectl wait --for=condition=ready pod -l app=rabbitmq --timeout=120s

# 5. Applications
kubectl apply -f k8s/30-user-service.yaml
kubectl apply -f k8s/31-order-service.yaml
kubectl apply -f k8s/32-payment-service.yaml
kubectl apply -f k8s/40-gateway.yaml

# 6. Ingress
kubectl apply -f k8s/50-ingress.yaml

# 7. Run migrations (Job)
kubectl apply -f k8s/60-migrations.yaml
```

---

### **Phase 7: Makefile Commands**

Tambahkan ke `Makefile`:
```makefile
# Kubernetes Commands
k8s-setup:
	./scripts/setup-k3d.sh

k8s-build:
	./scripts/build-images.sh

k8s-deploy:
	kubectl apply -f k8s/

k8s-delete:
	k3d cluster delete microservices

k8s-logs:
	kubectl logs -f deployment/gateway -n microservices

k8s-status:
	kubectl get pods -n microservices
```

---

### **Phase 8: Monitoring & Debugging**

**Tools:**
- `kubectl` commands untuk debugging
- `k9s` (TUI) - optional tapi recommended
- `stern` untuk log aggregation

**Health Checks:**
```bash
# Check all pods
kubectl get pods -n microservices

# Check services
kubectl get svc -n microservices

# Port-forward untuk akses langsung
kubectl port-forward svc/gateway 4000:4000 -n microservices

# Exec ke container
kubectl exec -it deployment/user-service -n microservices -- sh
```

---

### **Kebutuhan Resource Laptop**

**Minimum:**
- CPU: 4 cores
- RAM: 8GB
- Disk: 10GB free

**Recommended:**
- CPU: 6+ cores  
- RAM: 12GB+
- SSD untuk better I/O

---

### **Total Estimasi Waktu Setup**

| Task | Estimasi |
|------|----------|
| Install k3d | 2 menit |
| Build images | 3-5 menit |
| Deploy infrastructure | 2 menit |
| Deploy services | 1 menit |
| Verifikasi & test | 2 menit |
| **Total** | **~10-15 menit
