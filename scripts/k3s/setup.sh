#!/bin/bash
#
# k3s Setup Script
# Verifies prerequisites and prepares the environment
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K3S_SCRIPTS="${SCRIPT_DIR}"

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

check_prerequisites() {
    header "Step 1/4: Checking Prerequisites"
    
    local missing=()
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        missing+=("kubectl")
    else
        log_success "kubectl installed"
    fi
    
    # Check k3s
    if ! systemctl is-active --quiet k3s 2>/dev/null; then
        log_error "k3s is not running"
        echo "  Install: curl -sfL https://get.k3s.io | sh -"
        exit 1
    else
        log_success "k3s is running"
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing+=("go")
    else
        log_success "Go installed: $(go version | awk '{print $3}')"
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        missing+=("docker")
    else
        log_success "Docker installed"
    fi
    
    if [ ${#missing[@]} -ne 0 ]; then
        log_error "Missing prerequisites: ${missing[*]}"
        exit 1
    fi
}

setup_kubeconfig() {
    header "Step 2/4: Setting up Kubeconfig"
    
    K3S_CONFIG="/etc/rancher/k3s/k3s.yaml"
    USER_CONFIG="$HOME/.kube/config"
    
    if [ -f "$USER_CONFIG" ]; then
        log_success "Kubeconfig already exists at $USER_CONFIG"
    elif [ -f "$K3S_CONFIG" ]; then
        log_info "Copying k3s config to $USER_CONFIG..."
        mkdir -p "$HOME/.kube"
        if sudo cp "$K3S_CONFIG" "$USER_CONFIG" 2>/dev/null; then
            sudo chown $(id -u):$(id -g) "$USER_CONFIG"
            chmod 600 "$USER_CONFIG"
            log_success "Kubeconfig copied"
        else
            log_error "Failed to copy kubeconfig. Run manually:"
            echo "  sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config"
            exit 1
        fi
    else
        log_error "k3s config not found"
        exit 1
    fi
    
    # Verify connection
    export KUBECONFIG="$USER_CONFIG"
    if kubectl get nodes &> /dev/null; then
        log_success "Connected to k3s"
        kubectl get nodes
    else
        log_error "Cannot connect to k3s"
        exit 1
    fi
}

create_namespace() {
    header "Step 3/4: Creating Namespace"
    
    if kubectl get namespace microservices &> /dev/null; then
        log_warning "Namespace 'microservices' already exists"
    else
        kubectl create namespace microservices
        log_success "Namespace 'microservices' created"
    fi
}

show_next_steps() {
    header "Step 4/4: Setup Complete!"
    
    log_success "Environment ready"
    echo ""
    echo "Next steps:"
    echo "  1. Build:  ./scripts/k3s/build.sh"
    echo "  2. Deploy: ./scripts/k3s/deploy.sh"
    echo ""
    echo "Or run all at once:"
    echo "  ./setup.sh"
    echo ""
}

main() {
    check_prerequisites
    setup_kubeconfig
    create_namespace
    show_next_steps
}

main "$@"
