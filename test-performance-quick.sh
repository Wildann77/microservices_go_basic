#!/bin/bash

# Quick API Performance Report Generator
# Usage: ./test-performance-report.sh

set -e

echo "========================================"
echo "   API PERFORMANCE REPORT"
echo "   Microservices Go Project"
echo "   Date: $(date)"
echo "========================================"
echo ""

# Check services
echo "1. SERVICE STATUS"
echo "-----------------"
curl -s http://localhost:8081/health | jq -r '"User Service: \(.status)"' 2>/dev/null || echo "User Service: DOWN"
curl -s http://localhost:8082/health | jq -r '"Order Service: \(.status)"' 2>/dev/null || echo "Order Service: DOWN"
curl -s http://localhost:8083/health | jq -r '"Payment Service: \(.status)"' 2>/dev/null || echo "Payment Service: DOWN"
curl -s -o /dev/null -w "Nginx: %{http_code}\n" http://localhost/ || echo "Nginx: DOWN"
curl -s -o /dev/null -w "Gateway: %{http_code}\n" http://localhost:4000/ || echo "Gateway: DOWN"
echo ""

# Test response times
echo "2. RESPONSE TIME TEST"
echo "---------------------"
echo "Testing 5 requests to each endpoint..."
echo ""

test_endpoint() {
    local name=$1
    local url=$2
    local total=0
    local count=5
    
    echo -n "$name: "
    
    for i in $(seq 1 $count); do
        local time=$(curl -s -o /dev/null -w "%{time_total}" "$url" 2>/dev/null)
        total=$(echo "$total + $time" | bc)
        echo -n "."
    done
    
    local avg=$(echo "scale=3; $total / $count * 1000" | bc)
    echo " ${avg}ms average"
}

test_endpoint "User Health    " "http://localhost:8081/health"
test_endpoint "Order Health   " "http://localhost:8082/health"
test_endpoint "Payment Health " "http://localhost:8083/health"
test_endpoint "Gateway        " "http://localhost:4000/"
test_endpoint "Nginx          " "http://localhost/"

echo ""

# Test Nginx caching
echo "3. NGINX CACHING PERFORMANCE"
echo "----------------------------"
echo "Testing cache hit vs miss..."
echo ""

# Clear cache first
echo "Clearing cache..."
make nginx-clear-cache > /dev/null 2>&1 || true

# First request (miss)
echo -n "First request (cache MISS):  "
time1=$(curl -s -o /dev/null -w "%{time_total}" http://localhost/)
echo "${time1}s"

# Second request (hit)
echo -n "Second request (cache HIT):  "
time2=$(curl -s -o /dev/null -w "%{time_total}" http://localhost/)
echo "${time2}s"

# Calculate improvement
if command -v bc >/dev/null 2>&1; then
    improvement=$(echo "scale=1; (($time1 - $time2) / $time1 * 100)" | bc)
    echo "Cache improvement: ${improvement}%"
else
    echo "Cache is active: X-Cache-Status: $(curl -s -I http://localhost/ | grep -i x-cache | cut -d: -f2 | tr -d ' ')"
fi

echo ""

# Test authenticated endpoint
echo "4. AUTHENTICATED ENDPOINT TEST"
echo "------------------------------"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiM2Y5OGM0ZDQtYjY1NS00NmRiLWEyOTYtZGFjOGEzN2VhMzIwIiwiZW1haWwiOiJ0ZXN0dXNlckBleGFtcGxlLmNvbSIsInJvbGUiOiJ1c2VyIiwiaXNzIjoibWljcm9zZXJ2aWNlcy1nbyJ9.51dbm34NxJN6KDbZj4lEc0lPTaTS6cAPU2-TPSsqgJU"

echo -n "User API (Direct):    "
time_direct=$(curl -s -o /dev/null -w "%{time_total}" \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8081/api/v1/users/me 2>/dev/null || echo "0")
echo "${time_direct}s"

echo -n "User API (via Nginx): "
time_nginx=$(curl -s -o /dev/null -w "%{time_total}" \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost/api/v1/users/me 2>/dev/null || echo "0")
echo "${time_nginx}s"

echo ""

# Quick load test
echo "5. QUICK LOAD TEST"
echo "------------------"
echo "100 requests, 10 concurrent..."
echo ""

echo "Direct Gateway:"
ab -n 100 -c 10 http://localhost:4000/ 2>&1 | grep -E "Requests per second:|Time per request:" | head -2

echo ""
echo "Via Nginx:"
ab -n 100 -c 10 http://localhost/ 2>&1 | grep -E "Requests per second:|Time per request:" | head -2

echo ""
echo "========================================"
echo "   END OF PERFORMANCE REPORT"
echo "========================================"
