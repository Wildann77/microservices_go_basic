#!/bin/bash
#
# k3s Redeploy Script
# Quick redeploy without rebuilding images
#

set -e

NAMESPACE="microservices"

# Set kubeconfig
export KUBECONFIG="$HOME/.kube/config"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

header() {
    echo ""
    echo "======================================"
    echo "$1"
    echo "======================================"
    echo ""
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Check connection
export KUBECONFIG="$HOME/.kube/config"
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Error: Not connected to k3s${NC}"
    exit 1
fi

header "k3s Redeploy"

echo "Options:"
echo "  1. Soft redeploy (restart pods only, keep data)"
echo "  2. Hard redeploy (delete all, fresh deploy, keep images)"
echo "  3. Full reset (delete namespace, need re-import)"
echo ""
read -p "Choose option (1-3): " choice

case $choice in
    1)
        echo ""
        log_info "Soft redeploy: Restarting all deployments..."
        kubectl rollout restart deployment/gateway -n ${NAMESPACE}
        kubectl rollout restart deployment/user-service -n ${NAMESPACE}
        kubectl rollout restart deployment/order-service -n ${NAMESPACE}
        kubectl rollout restart deployment/payment-service -n ${NAMESPACE}
        log_success "All deployments restarted"
        echo ""
        echo "Waiting for pods..."
        sleep 5
        kubectl get pods -n ${NAMESPACE}
        ;;
    
    2)
        echo ""
        log_warning "Hard redeploy: Deleting all resources except namespace..."
        
        # Delete deployments
        log_info "Deleting deployments..."
        kubectl delete deployment gateway -n ${NAMESPACE} --ignore-not-found=true
        kubectl delete deployment user-service -n ${NAMESPACE} --ignore-not-found=true
        kubectl delete deployment order-service -n ${NAMESPACE} --ignore-not-found=true
        kubectl delete deployment payment-service -n ${NAMESPACE} --ignore-not-found=true
        
        # Delete ingress
        log_info "Deleting ingress..."
        kubectl delete ingress gateway-ingress -n ${NAMESPACE} --ignore-not-found=true
        
        # Keep infrastructure (databases, redis, rabbitmq)
        log_info "Keeping infrastructure (databases, redis, rabbitmq)..."
        
        echo ""
        log_info "Re-deploying applications..."
        ./scripts/k3s/deploy.sh
        
        log_success "Redeploy complete!"
        ;;
    
    3)
        echo ""
        log_warning "Full reset: This will delete EVERYTHING including data!"
        read -p "Are you sure? (yes/no): " confirm
        
        if [ "$confirm" == "yes" ]; then
            ./scripts/k3s/cleanup.sh
            echo ""
            log_info "To restart from scratch, run:"
            echo "  ./setup.sh"
        else
            echo "Cancelled"
        fi
        ;;
    
    *)
        echo "Invalid option"
        exit 1
        ;;
esac
