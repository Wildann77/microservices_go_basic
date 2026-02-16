#!/bin/bash
#
# k3s Cleanup Script with Image Removal Option
# Removes all microservices resources and optionally Docker images
#

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

show_menu() {
    header "k3s Cleanup Options"
    
    echo "What do you want to delete?"
    echo ""
    echo "1. Resources only (keep images)"
    echo "   - Delete namespace and all resources"
    echo "   - Keep Docker images in k3s"
    echo "   - Fast cleanup"
    echo ""
    echo "2. Resources + Images (full cleanup)"
    echo "   - Delete namespace and all resources"
    echo "   - Delete Docker images from k3s"
    echo "   - Free up disk space"
    echo "   - Complete fresh start"
    echo ""
    echo "3. Cancel"
    echo ""
}

cleanup_resources() {
    header "Cleaning up Resources"
    
    if kubectl get namespace ${NAMESPACE} &> /dev/null; then
        log_info "Deleting namespace '${NAMESPACE}'..."
        kubectl delete namespace ${NAMESPACE} --wait=false
        log_success "Namespace deletion initiated"
        
        echo ""
        log_info "Waiting for cleanup..."
        sleep 2
        
        # Check if deleted
        if kubectl get namespace ${NAMESPACE} &> /dev/null; then
            log_warning "Namespace still terminating (this is normal)"
            echo "Run 'kubectl get ns' to check status"
        else
            log_success "Namespace deleted"
        fi
    else
        log_warning "Namespace '${NAMESPACE}' not found"
    fi
}

cleanup_images() {
    header "Cleaning up Docker Images"
    
    log_warning "This will delete Docker images from k3s containerd"
    read -p "Continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        log_info "Image cleanup cancelled"
        return
    fi
    
    log_info "Removing images from k3s..."
    
    # List of images to delete
    images=(
        "localhost:5000/gateway:latest"
        "localhost:5000/user-service:latest"
        "localhost:5000/order-service:latest"
        "localhost:5000/payment-service:latest"
    )
    
    for img in "${images[@]}"; do
        log_info "Removing $img..."
        if sudo k3s ctr images rm "$img" 2>/dev/null; then
            log_success "Removed $img"
        else
            log_warning "Failed to remove $img (may not exist)"
        fi
    done
    
    log_success "Image cleanup complete"
}

cleanup_local() {
    header "Cleaning up Local Build Artifacts"
    
    if [ -d "build" ]; then
        log_info "Removing build directory..."
        rm -rf build/
        log_success "Build directory removed"
    fi
    
    if [ -d "dist" ]; then
        log_info "Removing dist directory..."
        rm -rf dist/
        log_success "Dist directory removed"
    fi
    
    # Remove tar files if any
    if ls *.tar 1> /dev/null 2>&1; then
        log_info "Removing tar files..."
        rm -f *.tar
        log_success "Tar files removed"
    fi
    
    if ls *.tar.gz 1> /dev/null 2>&1; then
        log_info "Removing tar.gz files..."
        rm -f *.tar.gz
        log_success "Tar.gz files removed"
    fi
}

show_next_steps() {
    echo ""
    echo "======================================"
    echo "Cleanup Complete!"
    echo "======================================"
    echo ""
    
    if [ "$CLEANUP_IMAGES" == "true" ]; then
        echo -e "${YELLOW}Full cleanup performed:${NC}"
        echo "  ✓ All resources deleted"
        echo "  ✓ Docker images deleted"
        echo "  ✓ Local build files deleted"
        echo ""
        echo "To start fresh, run:"
        echo "  1. ./setup.sh (full setup from scratch)"
    else
        echo -e "${GREEN}Resources cleaned up:${NC}"
        echo "  ✓ All Kubernetes resources deleted"
        echo "  ✓ Local build files deleted"
        echo "  ✗ Docker images kept (for faster redeploy)"
        echo ""
        echo "To restart:"
        echo "  1. ./scripts/k3s/deploy.sh (quick redeploy)"
        echo ""
        echo "To delete images later:"
        echo "  sudo k3s ctr images rm localhost:5000/gateway:latest"
        echo "  sudo k3s ctr images rm localhost:5000/user-service:latest"
        echo "  sudo k3s ctr images rm localhost:5000/order-service:latest"
        echo "  sudo k3s ctr images rm localhost:5000/payment-service:latest"
    fi
    echo ""
}

# Main
main() {
    show_menu
    
    read -p "Select option (1-3): " choice
    
    case $choice in
        1)
            CLEANUP_IMAGES="false"
            cleanup_resources
            cleanup_local
            show_next_steps
            ;;
        2)
            CLEANUP_IMAGES="true"
            cleanup_resources
            cleanup_images
            cleanup_local
            show_next_steps
            ;;
        3)
            log_info "Cancelled"
            exit 0
            ;;
        *)
            log_error "Invalid option"
            exit 1
            ;;
    esac
}

main "$@"
