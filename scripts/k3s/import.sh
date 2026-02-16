#!/bin/bash
#
# k3s Import Script
# Imports Docker images into k3s containerd
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../" && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"

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

import_images() {
    header "Importing Images to k3s"
    
    if [ ! -d "${BUILD_DIR}/images" ]; then
        log_error "Images not found. Run: ./scripts/k3s/build.sh"
        exit 1
    fi
    
    log_info "Importing gateway..."
    if sudo k3s ctr images import "${BUILD_DIR}/images/gateway.tar"; then
        log_success "gateway"
    else
        log_error "Failed to import gateway"
        exit 1
    fi
    
    log_info "Importing user-service..."
    if sudo k3s ctr images import "${BUILD_DIR}/images/user-service.tar"; then
        log_success "user-service"
    else
        log_error "Failed to import user-service"
        exit 1
    fi
    
    log_info "Importing order-service..."
    if sudo k3s ctr images import "${BUILD_DIR}/images/order-service.tar"; then
        log_success "order-service"
    else
        log_error "Failed to import order-service"
        exit 1
    fi
    
    log_info "Importing payment-service..."
    if sudo k3s ctr images import "${BUILD_DIR}/images/payment-service.tar"; then
        log_success "payment-service"
    else
        log_error "Failed to import payment-service"
        exit 1
    fi
}

verify_import() {
    header "Verifying Import"
    
    log_info "Images in k3s:"
    sudo k3s ctr images list 2>/dev/null | grep localhost | awk '{print "  - " $1}' || true
}

show_next_steps() {
    header "Import Complete!"
    
    echo "Next: Deploy to Kubernetes"
    echo "  ./scripts/k3s/deploy.sh"
    echo ""
}

main() {
    import_images
    verify_import
    show_next_steps
}

main "$@"
