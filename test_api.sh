#!/bin/bash

echo "Testing Minimax API with provided key..."

# Minimax API details
API_KEY="sk-cp-zfWlPgSp4TJZjG3y48kqD8dfUfdwXLGDT7HWBbbvGj568o9aODNaLYsQ72Gqe-HMiYlJigllQnUHrxzpFE6HI4Dioxw_SYTFUEr8jvnMFHWIuFBHg6VjqXQ"
BASE_URL="https://api.minimaxi.com/anthropic"
MODEL="MiniMax-M2.1"

echo "Testing endpoint: $BASE_URL/messages"
echo "Model: $MODEL"
echo "Sending test request..."

# Test the API
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE_URL/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"'"$MODEL"'","messages":[{"role":"user","content":"Hello, this is a test."}],"max_tokens":100}')

http_code=$(echo "$response" | tail -n1 | cut -d':' -f2)
response_body=$(echo "$response" | sed '$d')

echo "HTTP Status Code: $http_code"

if [ "$http_code" -eq 200 ]; then
    echo "✅ SUCCESS: API is working!"
    echo "Response: $response_body"
elif [ "$http_code" -eq 401 ]; then
    echo "❌ ERROR 401: Unauthorized - Invalid API key"
    echo "Response: $response_body"
elif [ "$http_code" -eq 404 ]; then
    echo "❌ ERROR 404: Endpoint not found - Checking alternative endpoints..."
    
    # Try with v1 suffix
    echo "Trying with /v1/messages endpoint..."
    response_v1=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "https://api.minimaxi.com/v1/messages" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $API_KEY" \
      -H "anthropic-version: 2023-06-01" \
      -d '{"model":"'"$MODEL"'","messages":[{"role":"user","content":"Hello, this is a test."}],"max_tokens":100}')
    
    http_code_v1=$(echo "$response_v1" | tail -n1 | cut -d':' -f2)
    response_body_v1=$(echo "$response_v1" | sed '$d')
    
    echo "HTTP Status Code for /v1/messages: $http_code_v1"
    
    if [ "$http_code_v1" -eq 200 ]; then
        echo "✅ SUCCESS: /v1/messages endpoint is working!"
        echo "Response: $response_body_v1"
        echo "API_ENDPOINT=https://api.minimaxi.com/v1/messages" > api_test_results.txt
    else
        echo "❌ ERROR: /v1/messages also failed"
        echo "Response: $response_body_v1"
        echo "API_ENDPOINT=FAILED" > api_test_results.txt
    fi
else
    echo "❌ ERROR: Other error occurred"
    echo "Response: $response_body"
    echo "API_ENDPOINT=FAILED" > api_test_results.txt
fi

echo "Test completed."