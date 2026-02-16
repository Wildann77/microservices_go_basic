#!/bin/bash
#
# k3s Logs Script
# View logs for services
#

NAMESPACE="microservices"

# Set kubeconfig
export KUBECONFIG="$HOME/.kube/config"

show_usage() {
    echo "Usage: $0 [service]"
    echo ""
    echo "Services:"
    echo "  gateway          - Gateway logs"
    echo "  user             - User service logs"
    echo "  order            - Order service logs"
    echo "  payment          - Payment service logs"
    echo "  all              - All services logs"
    echo ""
    echo "Examples:"
    echo "  $0 gateway       # View gateway logs"
    echo "  $0 user          # View user service logs"
    echo "  $0 all           # View all service logs"
}

view_logs() {
    local service=$1
    echo "=== Logs for $service ==="
    echo "(Press Ctrl+C to exit)"
    echo ""
    kubectl logs -f deployment/$service -n ${NAMESPACE}
}

view_all_logs() {
    echo "=== All Service Logs ==="
    echo "(Press Ctrl+C to exit)"
    echo ""
    kubectl logs -f --all-containers=true --selector='app in (gateway, user-service, order-service, payment-service)' -n ${NAMESPACE}
}

main() {
    if [ $# -eq 0 ]; then
        show_usage
        exit 1
    fi
    
    case "$1" in
        gateway)
            view_logs "gateway"
            ;;
        user|user-service)
            view_logs "user-service"
            ;;
        order|order-service)
            view_logs "order-service"
            ;;
        payment|payment-service)
            view_logs "payment-service"
            ;;
        all)
            view_all_logs
            ;;
        *)
            echo "Unknown service: $1"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
