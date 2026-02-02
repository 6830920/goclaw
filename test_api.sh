#!/bin/bash

echo "Testing OpenClaw-Go API endpoints..."

echo -e "\n1. Testing health endpoint:"
curl -s -X GET http://localhost:18888/health | jq .

echo -e "\n2. Testing root endpoint:"
curl -s -X GET http://localhost:18888/ | jq .

echo -e "\n3. Testing chat endpoint:"
curl -s -X POST http://localhost:18888/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?", "sessionId": "test123"}' | jq .

echo -e "\n4. Testing sessions endpoint:"
curl -s -X GET http://localhost:18888/api/sessions | jq .

echo -e "\n5. Testing memory stats endpoint:"
curl -s -X GET http://localhost:18888/api/memory/stats | jq .

echo -e "\nAPI tests completed!"