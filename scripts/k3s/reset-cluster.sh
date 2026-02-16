#!/bin/bash
#
# k3s Cluster Reset Script
# Resets k3s cluster to factory fresh state
#

# Set kubeconfig
export KUBECONFIG="$HOME/.kube/config"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

header() {
    echo ""
    echo -e "${BOLD}======================================${NC}"
    echo -e "${BOLD}$1${NC}"
    echo -e "${BOLD}======================================${NC}"
    echo ""
}

warning() {
    echo -e "${RED}WARNING: This will completely reset your k3s cluster!${NC}"
    echo ""
    echo "This will delete:"
    echo "  ✗ All namespaces (not just microservices)"
    echo "  ✗ All pods, deployments, services"
    echo "  ✗ All Docker images in k3s"
    echo "  ✗ All persistent data"
    echo "  ✗ All configurations"
    echo ""
    echo "Your k3s cluster will be like a fresh installation."
    echo ""
}

show_menu() {
    header "k3s Cluster Reset Options"
    
    echo "What do you want to do?"
    echo ""
    echo "1. ${YELLOW}Soft Reset${NC} - Delete all resources but keep cluster"
    echo "   Delete: All namespaces, pods, deployments"
    echo "   Keep: k3s cluster running"
    echo "   Time: ~30 seconds"
    echo ""
    echo "2. ${RED}Hard Reset${NC} - Factory reset k3s cluster"
    echo "   Delete: Everything + k3s data"
    echo "   Result: Fresh k3s installation"
    echo "   Time: ~1-2 minutes"
    echo "   Note: k3s will auto-restart"
    echo ""
    echo "3. ${RED}Complete Reinstall${NC} - Uninstall & reinstall k3s"
    echo "   Delete: k3s completely"
    echo "   Reinstall: Fresh k3s"
    echo "   Time: ~2-3 minutes"
    echo ""
    echo "4. Cancel"
    echo ""
}

soft_reset() {
    header "Performing Soft Reset"
    
    echo -e "${YELLOW}Deleting all namespaces...${NC}"
    
    # Delete all namespaces except system ones
    kubectl get namespaces --no-headers | grep -v "kube-system\|kube-public\|kube-node-lease\|default" | awk '{print $1}' | while read ns; do
        echo "Deleting namespace: $ns"
        kubectl delete namespace "$ns" --wait=false 2>/dev/null || true
    done
    
    # Delete all resources in default namespace
    echo "Cleaning default namespace..."
    kubectl delete all --all -n default 2>/dev/null || true
    
    # Delete all images
    echo ""
    echo -e "${YELLOW}Removing all Docker images...${NC}"
    sudo k3s ctr images ls -q | xargs -I {} sudo k3s ctr images rm {} 2>/dev/null || true
    
    echo ""
    echo -e "${GREEN}✓ Soft reset complete!${NC}"
    echo ""
    echo "k3s cluster is still running but empty."
    echo "Run './setup.sh' to deploy microservices again."
}

hard_reset() {
    header "Performing Hard Reset"
    
    echo -e "${RED}Stopping k3s...${NC}"
    sudo systemctl stop k3s
    
    echo -e "${YELLOW}Cleaning k3s data...${NC}"
    sudo rm -rf /var/lib/rancher/k3s/
    sudo rm -rf /etc/rancher/k3s/
    sudo rm -rf ~/.kube/config
    sudo rm -rf ~/.kube/cache
    
    echo -e "${GREEN}Starting k3s...${NC}"
    sudo systemctl start k3s
    
    echo ""
    echo -e "${GREEN}✓ Hard reset complete!${NC}"
    echo ""
    echo "k3s cluster has been reset to factory defaults."
    echo ""
    echo "Next steps:"
    echo "  1. Wait for k3s to be ready: sudo systemctl status k3s"
    echo "  2. Setup kubeconfig: ./scripts/k3s/setup.sh"
    echo "  3. Deploy: ./setup.sh"
}

complete_reinstall() {
    header "Complete k3s Reinstall"
    
    echo -e "${RED}Uninstalling k3s...${NC}"
    /usr/local/bin/k3s-uninstall.sh 2>/dev/null || sudo /usr/local/bin/k3s-uninstall.sh
    
    echo ""
    echo -e "${GREEN}Installing fresh k3s...${NC}"
    curl -sfL https://get.k3s.io | sh -
    
    echo ""
    echo -e "${GREEN}✓ k3s reinstalled!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Wait for k3s: sudo systemctl status k3s"
    echo "  2. Run full setup: ./setup.sh"
}

# Main
main() {
    warning
    show_menu
    
    read -p "Select option (1-4): " choice
    
    case $choice in
        1)
            read -p "Are you sure? (yes/no): " confirm
            if [ "$confirm" == "yes" ]; then
                soft_reset
            else
                echo "Cancelled"
            fi
            ;;
        2)
            read -p "Are you ABSOLUTELY sure? This will delete ALL data! (yes/no): " confirm
            if [ "$confirm" == "yes" ]; then
                hard_reset
            else
                echo "Cancelled"
            fi
            ;;
        3)
            read -p "This will completely remove k3s. Continue? (yes/no): " confirm
            if [ "$confirm" == "yes" ]; then
                complete_reinstall
            else
                echo "Cancelled"
            fi
            ;;
        4)
            echo "Cancelled"
            exit 0
            ;;
        *)
            echo "Invalid option"
            exit 1
            ;;
    esac
}

main "$@"
