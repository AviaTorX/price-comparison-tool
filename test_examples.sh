#!/bin/bash

echo "=== Price Comparison Tool - Test Examples ==="
echo ""

# Required test case - iPhone 16 Pro, 128GB in US
echo "1. Testing required example: iPhone 16 Pro, 128GB in US"
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{"country": "US", "query": "iPhone 16 Pro, 128GB"}' \
  | jq '.'

echo ""
echo "----------------------------------------"
echo ""

# Test case 2 - Indian market example
echo "2. Testing Indian market: boAt Airdopes 311 Pro"
curl -X POST http://localhost:8080/api/v1/prices \
  -H "Content-Type: application/json" \
  -d '{"country": "IN", "query": "boAt Airdopes 311 Pro"}' \
  | jq '.'

echo ""
echo "----------------------------------------"
echo ""

# Health check
echo "3. Health check"
curl -X GET http://localhost:8080/api/v1/health | jq '.'

echo ""
echo "----------------------------------------"
echo ""

# Supported sites
echo "4. Supported sites"
curl -X GET http://localhost:8080/api/v1/sites | jq '.'

echo ""
echo "=== All tests completed ==="