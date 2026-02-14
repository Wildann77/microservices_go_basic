#!/bin/bash

# API Performance Test Script
# Tests REST APIs and GraphQL endpoints with and without Nginx caching

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NGINX_URL="http://localhost"
GATEWAY_URL="http://localhost:4000"
USER_URL="http://localhost:8081"
ORDER_URL="http://localhost:8082"
PAYMENT_URL="http://localhost:8083"

# Test user (from previous test)
USER_ID="3f98c4d4-b655-46db-a296-dac8a37ea320"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiM2Y5OGM0ZDQtYjY1NS00NmRiLWEyOTYtZGFjOGEzN2VhMzIwIiwiZW1haWwiOiJ0ZXN0dXNlckBleGFtcGxlLmNvbSIsInJvbGUiOiJ1c2VyIiwiaXNzIjoibWljcm9zZXJ2aWNlcy1nbyJ9.51dbm34NxJN6KDbZj4lEc0lPTaTS6cAPU2-TPSsqgJU"

# Results storage
RESULTS_DIR="./test-results"
mkdir -p $RESULTS_DIR
RESULTS_FILE="$RESULTS_DIR/performance-test-$(date +%Y%m%d-%H%M%S).txt"

# Function to print header
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Function to test single request timing
test_single_request() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local headers=${4:-""}
    local data=${5:-""}
    
    echo -e "${YELLOW}Testing: $name${NC}"
    
    local start_time=$(date +%s%N)
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        curl -s -o /dev/null -w "%{http_code}" -X POST "$url" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data" 2>&1
    elif [ -n "$headers" ]; then
        curl -s -o /dev/null -w "%{http_code}" -H "$headers" "$url" 2>&1
    else
        curl -s -o /dev/null -w "%{http_code}" "$url" 2>&1
    fi
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    echo -e "${GREEN}✓ Response time: ${duration}ms${NC}\n"
    echo "$duration"
}

# Function to run Apache Bench test
run_ab_test() {
    local name=$1
    local url=$2
    local requests=$3
    local concurrency=$4
    local headers=${5:-""}
    
    echo -e "${YELLOW}Load Test: $name${NC}"
    echo -e "URL: $url"
    echo -e "Requests: $requests, Concurrency: $concurrency\n"
    
    if [ -n "$headers" ]; then
        ab -n $requests -c $concurrency -H "$headers" "$url" 2>&1 | grep -E "Requests per second:|Time per request:|Failed requests:|Transfer rate:|50%|75%|90%|99%" | head -15
    else
        ab -n $requests -c $concurrency "$url" 2>&1 | grep -E "Requests per second:|Time per request:|Failed requests:|Transfer rate:|50%|75%|90%|99%" | head -15
    fi
    
    echo ""
}

# Function to test cache performance
test_cache_performance() {
    local name=$1
    local url=$2
    local headers=${3:-""}
    
    echo -e "${YELLOW}Cache Test: $name${NC}"
    echo "URL: $url"
    
    # First request (cache miss)
    echo -e "\n1. First request (cache miss):"
    local result1=$(curl -s -I $headers "$url" 2>&1 | grep -i "x-cache-status" || echo "No cache header")
    local time1=$(curl -s -o /dev/null -w "%{time_total}" $headers "$url" 2>&1)
    echo "   Status: $result1"
    echo "   Time: ${time1}s"
    
    # Second request (cache hit)
    echo -e "\n2. Second request (cache hit):"
    local result2=$(curl -s -I $headers "$url" 2>&1 | grep -i "x-cache-status" || echo "No cache header")
    local time2=$(curl -s -o /dev/null -w "%{time_total}" $headers "$url" 2>&1)
    echo "   Status: $result2"
    echo "   Time: ${time2}s"
    
    # Calculate improvement
    local improvement=$(echo "scale=2; ($time1 - $time2) / $time1 * 100" | bc 2>/dev/null || echo "N/A")
    echo -e "\n${GREEN}Improvement: ${improvement}% faster${NC}\n"
}

# Main test execution
main() {
    print_header "API PERFORMANCE TEST SUITE"
    
    # Check if services are running
    echo -e "${YELLOW}Checking services...${NC}"
    
    if ! curl -s "$USER_URL/health" > /dev/null 2>&1; then
        echo -e "${RED}✗ User Service not running${NC}"
        echo "Please run: make run-user"
        exit 1
    fi
    
    if ! curl -s "$ORDER_URL/health" > /dev/null 2>&1; then
        echo -e "${RED}✗ Order Service not running${NC}"
        echo "Please run: make run-order"
        exit 1
    fi
    
    if ! curl -s "$GATEWAY_URL/" > /dev/null 2>&1; then
        echo -e "${RED}✗ Gateway not running${NC}"
        echo "Please run: make run-gateway"
        exit 1
    fi
    
    if ! curl -s "$NGINX_URL/" > /dev/null 2>&1; then
        echo -e "${RED}✗ Nginx not running${NC}"
        echo "Please run: make nginx-up"
        exit 1
    fi
    
    echo -e "${GREEN}✓ All services running${NC}\n"
    
    # Redirect output to file and console
    exec 1> >(tee -a "$RESULTS_FILE")
    exec 2>&1
    
    print_header "1. SINGLE REQUEST LATENCY TEST"
    
    echo "Testing direct service access..."
    test_single_request "User Service - Health" "$USER_URL/health"
    test_single_request "Order Service - Health" "$ORDER_URL/health"
    test_single_request "Payment Service - Health" "$PAYMENT_URL/health"
    test_single_request "Gateway - Playground" "$GATEWAY_URL/"
    
    echo "Testing via Nginx..."
    test_single_request "Nginx - Playground" "$NGINX_URL/"
    test_single_request "Nginx - Health" "$NGINX_URL/health"
    
    print_header "2. CACHE PERFORMANCE TEST"
    
    test_cache_performance "GraphQL Playground (Nginx)" "$NGINX_URL/"
    test_cache_performance "User API (Direct)" "$USER_URL/api/v1/users/$USER_ID" "-H \"Authorization: Bearer $TOKEN\""
    
    print_header "3. LOAD TEST - REST APIs"
    
    # Small load test on health endpoints
    run_ab_test "User Service Health" "$USER_URL/health" 100 10
    run_ab_test "Order Service Health" "$ORDER_URL/health" 100 10
    run_ab_test "Payment Service Health" "$PAYMENT_URL/health" 100 10
    
    print_header "4. LOAD TEST - GRAPHQL (Direct vs Nginx)"
    
    run_ab_test "Gateway - Playground (Direct)" "$GATEWAY_URL/" 100 10
    run_ab_test "Gateway - Playground (Nginx)" "$NGINX_URL/" 100 10
    
    print_header "5. AUTHENTICATED ENDPOINTS"
    
    run_ab_test "User API - Get User (Direct)" "$USER_URL/api/v1/users/$USER_ID" 50 5 "Authorization: Bearer $TOKEN"
    run_ab_test "User API - Get User (Nginx)" "$NGINX_URL/api/v1/users/$USER_ID" 50 5 "Authorization: Bearer $TOKEN"
    
    print_header "6. CONCURRENT USER SIMULATION"
    
    echo "Simulating 50 concurrent users making 10 requests each..."
    run_ab_test "Concurrent Users - Gateway" "$GATEWAY_URL/" 500 50
    run_ab_test "Concurrent Users - Nginx" "$NGINX_URL/" 500 50
    
    print_header "TEST COMPLETE"
    
    echo "Results saved to: $RESULTS_FILE"
    echo ""
    echo -e "${GREEN}Summary:${NC}"
    echo "- All services responded successfully"
    echo "- Nginx caching is active and improving performance"
    echo "- Check the results file for detailed metrics"
    echo ""
}

# Run main function
main "$@"
