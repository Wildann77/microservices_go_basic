#!/bin/bash
#
# k3s Deploy Script
# Deploys all Kubernetes manifests
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../" && pwd)"
K8S_DIR="${PROJECT_ROOT}/k8s"
NAMESPACE="microservices"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

header() {
    echo ""
    echo "======================================"
    echo "$1"
    echo "======================================"
    echo ""
}

step() {
    echo ""
    echo -e "${YELLOW}[STEP $1]${NC} $2"
    echo ""
}

check_connection() {
    export KUBECONFIG="$HOME/.kube/config"
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Not connected to k3s"
        exit 1
    fi
}

deploy_infrastructure() {
    step "1/3" "Deploying Infrastructure"
    
    log_info "Creating namespace..."
    kubectl apply -f "${K8S_DIR}/00-namespace.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Applying ConfigMap..."
    kubectl apply -f "${K8S_DIR}/02-configmap.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Applying Secrets..."
    kubectl apply -f "${K8S_DIR}/03-secrets.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Deploying databases..."
    kubectl apply -f "${K8S_DIR}/10-postgres-user.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    kubectl apply -f "${K8S_DIR}/11-postgres-order.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    kubectl apply -f "${K8S_DIR}/12-postgres-payment.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Deploying message queue..."
    kubectl apply -f "${K8S_DIR}/20-rabbitmq.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Deploying cache..."
    kubectl apply -f "${K8S_DIR}/21-redis.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_success "Infrastructure deployed"
}

wait_for_infrastructure() {
    step "2/3" "Waiting for Infrastructure"
    
    echo "Waiting for PostgreSQL..."
    kubectl wait --for=condition=ready pod -l app=postgres-user -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "postgres-user timeout"
    kubectl wait --for=condition=ready pod -l app=postgres-order -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "postgres-order timeout"
    kubectl wait --for=condition=ready pod -l app=postgres-payment -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "postgres-payment timeout"
    
    echo "Waiting for Redis..."
    kubectl wait --for=condition=ready pod -l app=redis -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "redis timeout"
    
    echo "Waiting for RabbitMQ (may take 60-90s)..."
    kubectl wait --for=condition=ready pod -l app=rabbitmq -n ${NAMESPACE} --timeout=180s 2>/dev/null || log_warning "rabbitmq timeout"
    
    log_success "Infrastructure ready"
}

deploy_applications() {
    step "3/3" "Deploying Applications"
    
    log_info "Deploying user-service..."
    kubectl apply -f "${K8S_DIR}/30-user-service.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Deploying order-service..."
    kubectl apply -f "${K8S_DIR}/31-order-service.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Deploying payment-service..."
    kubectl apply -f "${K8S_DIR}/32-payment-service.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_info "Deploying gateway..."
    kubectl apply -f "${K8S_DIR}/40-gateway.yaml" 2>&1 | grep -E "(created|configured|unchanged)" || true
    
    log_success "Applications deployed"
}

wait_for_applications() {
    echo ""
    echo "Waiting for applications..."
    
    kubectl wait --for=condition=ready pod -l app=user-service -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "user-service timeout"
    kubectl wait --for=condition=ready pod -l app=order-service -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "order-service timeout"
    kubectl wait --for=condition=ready pod -l app=payment-service -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "payment-service timeout"
    kubectl wait --for=condition=ready pod -l app=gateway -n ${NAMESPACE} --timeout=120s 2>/dev/null || log_warning "gateway timeout"
}

show_summary() {
    header "Deployment Complete!"
    
    echo "Pod Status:"
    kubectl get pods -n ${NAMESPACE}
    
    echo ""
    echo "Services:"
    kubectl get svc -n ${NAMESPACE}
    
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null || echo "localhost")
    
    echo ""
    echo "======================================"
    echo "Access URLs:"
    echo "======================================"
    echo "  Gateway:     http://${NODE_IP}:30080"
    echo "  Health:      http://${NODE_IP}:30080/health"
    echo "  GraphQL:     http://${NODE_IP}:30080/"
    echo ""
    echo "Commands:"
    echo "  ./scripts/k3s/status.sh    # Check status"
    echo "  ./scripts/k3s/logs.sh      # View logs"
    echo "  ./scripts/k3s/cleanup.sh   # Clean up"
}

main() {
    check_connection
    deploy_infrastructure
    wait_for_infrastructure
    deploy_applications
    wait_for_applications
    show_summary
}

main "$@"
