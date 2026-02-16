#!/bin/bash
#
# k3s Test Script
# Comprehensive testing for k3s deployment
#

set -e

NAMESPACE="microservices"

# Set kubeconfig
export KUBECONFIG="${HOME}/.kube/config"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

header() {
    echo ""
    echo -e "${BOLD}======================================${NC}"
    echo -e "${BOLD}$1${NC}"
    echo -e "${BOLD}======================================${NC}"
    echo ""
}

# Test 1: Check k3s connection
test_k3s_connection() {
    header "Test 1: k3s Connection"
    
    export KUBECONFIG="${KUBECONFIG}"
    
    if kubectl cluster-info &> /dev/null; then
        log_success "Connected to k3s"
        kubectl version --short 2>/dev/null | grep "Server Version" | awk '{print "  Server: " $3}'
    else
        log_error "Cannot connect to k3s"
        echo "  Fix: export KUBECONFIG=${KUBECONFIG}"
        return 1
    fi
}

# Test 2: Check namespace exists
test_namespace() {
    header "Test 2: Namespace"
    
    if kubectl get namespace ${NAMESPACE} &> /dev/null; then
        log_success "Namespace '${NAMESPACE}' exists"
    else
        log_error "Namespace '${NAMESPACE}' not found"
        return 1
    fi
}

# Test 3: Check all pods are running
test_pods() {
    header "Test 3: Pod Status"
    
    local failed_pods=()
    
    # Get all pods
    while IFS= read -r line; do
        local pod=$(echo "$line" | awk '{print $1}')
        local ready=$(echo "$line" | awk '{print $2}')
        local status=$(echo "$line" | awk '{print $3}')
        
        if [[ "$status" == "Running" && "$ready" == *"/"* ]]; then
            local ready_count=$(echo "$ready" | cut -d'/' -f1)
            local total_count=$(echo "$ready" | cut -d'/' -f2)
            
            if [ "$ready_count" -eq "$total_count" ]; then
                log_success "$pod is Running ($ready)"
            else
                log_warning "$pod is Running but not fully ready ($ready)"
            fi
        else
            log_error "$pod is not Running (Status: $status, Ready: $ready)"
            failed_pods+=("$pod")
        fi
    done < <(kubectl get pods -n ${NAMESPACE} --no-headers 2>/dev/null)
    
    if [ ${#failed_pods[@]} -gt 0 ]; then
        echo ""
        echo "Failed pods:"
        for pod in "${failed_pods[@]}"; do
            echo "  - $pod"
            kubectl logs "$pod" -n ${NAMESPACE} --tail=5 2>/dev/null || true
        done
        return 1
    fi
}

# Test 4: Check services
test_services() {
    header "Test 4: Services"
    
    local services=("gateway" "user-service" "order-service" "payment-service" "postgres-user" "postgres-order" "postgres-payment" "rabbitmq" "redis")
    
    for svc in "${services[@]}"; do
        if kubectl get svc "$svc" -n ${NAMESPACE} &> /dev/null; then
            local cluster_ip=$(kubectl get svc "$svc" -n ${NAMESPACE} -o jsonpath='{.spec.clusterIP}')
            log_success "$svc service exists (ClusterIP: $cluster_ip)"
        else
            log_error "$svc service not found"
        fi
    done
}

# Test 5: Gateway Health Check
test_gateway_health() {
    header "Test 5: Gateway Health"
    
    # Test via NodePort
    local node_port_url="http://localhost:30080"
    
    log_info "Testing via NodePort ($node_port_url)..."
    if curl -s "${node_port_url}/health" | grep -q "healthy"; then
        log_success "Gateway health check via NodePort"
    else
        log_error "Gateway health check via NodePort failed"
    fi
    
    # Test via Ingress (if configured)
    log_info "Testing via Ingress..."
    local node_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
    if curl -s -H "Host: microservices.local" "http://${node_ip}/health" 2>/dev/null | grep -q "healthy"; then
        log_success "Gateway health check via Ingress"
    else
        log_warning "Gateway via Ingress not accessible (may need /etc/hosts entry)"
    fi
}

# Test 6: Database connectivity
test_databases() {
    header "Test 6: Database Connectivity"
    
    local dbs=("postgres-user" "postgres-order" "postgres-payment")
    
    for db in "${dbs[@]}"; do
        log_info "Checking $db..."
        if kubectl exec "$db-0" -n ${NAMESPACE} -- pg_isready -U postgres &> /dev/null; then
            log_success "$db is ready"
        else
            log_error "$db is not ready"
        fi
    done
}

# Test 7: Redis connectivity
test_redis() {
    header "Test 7: Redis Connectivity"
    
    log_info "Checking Redis..."
    if kubectl exec deployment/redis -n ${NAMESPACE} -- redis-cli ping 2>/dev/null | grep -q "PONG"; then
        log_success "Redis is responding"
    else
        log_error "Redis is not responding"
    fi
}

# Test 8: RabbitMQ connectivity
test_rabbitmq() {
    header "Test 8: RabbitMQ Connectivity"
    
    log_info "Checking RabbitMQ..."
    if kubectl exec rabbitmq-0 -n ${NAMESPACE} -- rabbitmq-diagnostics ping 2>/dev/null | grep -q "Health check passed"; then
        log_success "RabbitMQ is responding"
    else
        log_warning "RabbitMQ check failed (may still be starting up)"
    fi
}

# Test 9: Ingress configuration
test_ingress() {
    header "Test 9: Ingress Configuration"
    
    if kubectl get ingress gateway-ingress -n ${NAMESPACE} &> /dev/null; then
        log_success "Ingress 'gateway-ingress' exists"
        
        local host=$(kubectl get ingress gateway-ingress -n ${NAMESPACE} -o jsonpath='{.spec.rules[0].host}')
        echo "  Host: $host"
        
        local backend=$(kubectl get ingress gateway-ingress -n ${NAMESPACE} -o jsonpath='{.spec.rules[0].http.paths[0].backend.service.name}')
        echo "  Backend: $backend"
    else
        log_error "Ingress 'gateway-ingress' not found"
        echo "  Fix: kubectl apply -f k8s/50-ingress.yaml"
    fi
}

# Test 10: Resource usage
test_resources() {
    header "Test 10: Resource Usage"
    
    echo "Pod Resource Usage:"
    kubectl top pods -n ${NAMESPACE} 2>/dev/null || log_warning "Metrics not available (requires metrics-server)"
}

# Summary
show_summary() {
    header "Test Summary"
    
    echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}${BOLD}✓ All tests passed!${NC}"
        echo ""
        echo "Your k3s deployment is working correctly."
        echo "Access your application at:"
        echo "  http://localhost:30080 (NodePort)"
        echo "  http://microservices.local (Ingress - requires /etc/hosts)"
        return 0
    else
        echo -e "${RED}${BOLD}✗ Some tests failed${NC}"
        echo ""
        echo "Troubleshooting:"
        echo "  1. Check pod logs: kubectl logs <pod-name> -n ${NAMESPACE}"
        echo "  2. Check events: kubectl get events -n ${NAMESPACE}"
        echo "  3. Restart deployment: kubectl rollout restart deployment/<name> -n ${NAMESPACE}"
        return 1
    fi
}

# Main
main() {
    echo -e "${BOLD}"
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║     k3s Microservices Deployment Test Suite           ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    test_k3s_connection || exit 1
    test_namespace || exit 1
    test_pods
    test_services
    test_gateway_health
    test_databases
    test_redis
    test_rabbitmq
    test_ingress
    test_resources
    
    show_summary
}

main "$@"
