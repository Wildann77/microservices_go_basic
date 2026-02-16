#!/bin/bash
#
# k3s Status Script
# Shows status of all microservices
#

NAMESPACE="microservices"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "======================================"
echo "Microservices Status"
echo "======================================"
echo ""

# Check if namespace exists
if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo -e "${RED}Namespace '${NAMESPACE}' not found${NC}"
    echo "Run: ./scripts/k3s/setup.sh"
    exit 1
fi

echo "Pods:"
kubectl get pods -n ${NAMESPACE}

echo ""
echo "Services:"
kubectl get svc -n ${NAMESPACE}

echo ""
echo "Deployments:"
kubectl get deployments -n ${NAMESPACE}

echo ""
echo "======================================"
echo "Health Checks"
echo "======================================"
echo ""

# Test gateway health
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null || echo "localhost")

if curl -s http://localhost:30080/health &> /dev/null; then
    echo -e "${GREEN}✓${NC} Gateway: Healthy"
else
    echo -e "${RED}✗${NC} Gateway: Not responding"
fi

echo ""
echo "Quick Links:"
echo ""
echo "NodePort (Direct):"
echo "  Gateway:     http://${NODE_IP}:30080"
echo "  Health:      http://${NODE_IP}:30080/health"
echo "  GraphQL:     http://${NODE_IP}:30080/query"
echo ""
echo "Ingress (without host):"
echo "  Gateway:     http://${NODE_IP}"
echo "  Health:      http://${NODE_IP}/health"
echo "  GraphQL:     http://${NODE_IP}/query"
echo ""
echo "Ingress (with host - add to /etc/hosts):"
echo "  127.0.0.1 microservices.local"
echo "  Gateway:     http://microservices.local"
echo "  Health:      http://microservices.local/health"
echo "  GraphQL:     http://microservices.local/query"
