#!/bin/bash
#
# k3s Build Script
# Builds binaries locally and creates Docker images
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../" && pwd)"
BUILD_DIR="${PROJECT_ROOT}/build"
DIST_DIR="${PROJECT_ROOT}/dist"

# Set kubeconfig (for checking connection)
export KUBECONFIG="$HOME/.kube/config"

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

check_kubeconfig() {
    export KUBECONFIG="$HOME/.kube/config"
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Not connected to k3s. Run: ./scripts/k3s/setup.sh"
        exit 1
    fi
}

build_binaries() {
    step "1/4" "Building Go Binaries"
    
    mkdir -p "${DIST_DIR}"
    cd "${PROJECT_ROOT}"
    
    # Gateway
    log_info "Building gateway..."
    cd gateway
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -extldflags '-static'" \
        -o "${DIST_DIR}/gateway" \
        ./cmd/main.go 2>&1 | grep -v "^go: downloading" || true
    cd ..
    log_success "gateway"
    
    # User Service
    log_info "Building user-service..."
    cd services/user
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -extldflags '-static'" \
        -o "${DIST_DIR}/user-service" \
        ./cmd/main.go 2>&1 | grep -v "^go: downloading" || true
    cd ../..
    log_success "user-service"
    
    # Order Service
    log_info "Building order-service..."
    cd services/order
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -extldflags '-static'" \
        -o "${DIST_DIR}/order-service" \
        ./cmd/main.go 2>&1 | grep -v "^go: downloading" || true
    cd ../..
    log_success "order-service"
    
    # Payment Service
    log_info "Building payment-service..."
    cd services/payment
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -extldflags '-static'" \
        -o "${DIST_DIR}/payment-service" \
        ./cmd/main.go 2>&1 | grep -v "^go: downloading" || true
    cd ../..
    log_success "payment-service"
}

copy_migrations() {
    step "2/4" "Copying Migrations"
    
    cp -r services/user/migrations "${DIST_DIR}/"
    log_success "user migrations"
    
    cp -r services/order/migrations "${DIST_DIR}/"
    log_success "order migrations"
    
    cp -r services/payment/migrations "${DIST_DIR}/"
    log_success "payment migrations"
}

build_images() {
    step "3/4" "Creating Docker Images"
    
    cd "${DIST_DIR}"
    
    # Create Dockerfile for services
    cat > Dockerfile.service << 'DOCKERFILE'
FROM scratch
COPY binary /app/service
COPY migrations /app/migrations
WORKDIR /app
EXPOSE 8080
CMD ["/app/service"]
DOCKERFILE
    
    # Create Dockerfile for gateway
    cat > Dockerfile.gateway << 'DOCKERFILE'
FROM scratch
COPY gateway /app/gateway
WORKDIR /app
EXPOSE 4000
CMD ["/app/gateway"]
DOCKERFILE
    
    # Build gateway
    log_info "Building gateway image..."
    docker build -f Dockerfile.gateway -t localhost:5000/gateway:latest . 2>&1 | tail -1
    log_success "gateway:latest"
    
    # Build user-service
    log_info "Building user-service image..."
    cp user-service binary
    cp -r migrations/user migrations
    docker build -f Dockerfile.service -t localhost:5000/user-service:latest . 2>&1 | tail -1
    log_success "user-service:latest"
    
    # Build order-service
    log_info "Building order-service image..."
    cp order-service binary
    cp -r migrations/order migrations
    docker build -f Dockerfile.service -t localhost:5000/order-service:latest . 2>&1 | tail -1
    log_success "order-service:latest"
    
    # Build payment-service
    log_info "Building payment-service image..."
    cp payment-service binary
    cp -r migrations/payment migrations
    docker build -f Dockerfile.service -t localhost:5000/payment-service:latest . 2>&1 | tail -1
    log_success "payment-service:latest"
}

save_images() {
    step "4/4" "Saving Images"
    
    mkdir -p "${BUILD_DIR}/images"
    
    log_info "Saving images to tar files..."
    docker save localhost:5000/gateway:latest > "${BUILD_DIR}/images/gateway.tar"
    docker save localhost:5000/user-service:latest > "${BUILD_DIR}/images/user-service.tar"
    docker save localhost:5000/order-service:latest > "${BUILD_DIR}/images/order-service.tar"
    docker save localhost:5000/payment-service:latest > "${BUILD_DIR}/images/payment-service.tar"
    
    log_success "All images saved"
    ls -lh "${BUILD_DIR}/images/"
}

show_import_instructions() {
    header "Build Complete!"
    
    echo "Images are ready in: ${BUILD_DIR}/images/"
    echo ""
    echo "Next: Import images to k3s"
    echo "  sudo ./scripts/k3s/import.sh"
    echo ""
}

main() {
    check_kubeconfig
    build_binaries
    copy_migrations
    build_images
    save_images
    show_import_instructions
}

main "$@"
