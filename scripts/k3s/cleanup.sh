#!/bin/bash
#
# k3s Cleanup Script
# Removes all microservices resources
#

NAMESPACE="microservices"

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================"
echo "k3s Cleanup"
echo "======================================"
echo ""

if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo "Namespace '${NAMESPACE}' not found"
    exit 0
fi

echo -e "${YELLOW}WARNING: This will delete all microservices resources${NC}"
echo ""
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Cancelled"
    exit 0
fi

echo ""
echo "Deleting namespace..."
kubectl delete namespace ${NAMESPACE} --wait=false

echo ""
echo "Cleanup initiated!"
echo ""
echo "To verify:"
echo "  kubectl get pods -n ${NAMESPACE}"
echo ""
echo "Note: This does NOT delete:"
echo "  - k3s cluster"
echo "  - Docker images"
echo "  - Build artifacts"
