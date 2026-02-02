#!/bin/bash

# Test script to validate API keys and endpoints

echo "Testing Minimax API key and endpoint..."
API_KEY="sk-cp-zfWlPgSp4TJZjG3y48kqD8dfUfdwXLGDT7HWBbbvGj568o9aODNaLYsQ72Gqe-HMiYlJigllQnUHrxzpFE6HI4Dioxw_SYTFUEr8jvnMFHWIuFBHg6VjqXQ"

echo "Testing with CN endpoint: https://api.minimaxi.com/anthropic"
response=$(curl -s -w "\n%{http_code}" -X POST "https://api.minimaxi.com/anthropic/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"MiniMax-M2.1","messages":[{"role":"user","content":"你好"}],"max_tokens":100}')

http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | sed '$d')

echo "Response from CN endpoint:"
echo "Status: $http_code"
echo "Body: $response_body"
echo ""

if [ "$http_code" -eq 200 ]; then
    echo "✓ Minimax CN endpoint is working!"
else
    echo "✗ Minimax CN endpoint failed with status $http_code"
    
    echo "Testing with Global endpoint: https://api.minimax.io/anthropic"
    response=$(curl -s -w "\n%{http_code}" -X POST "https://api.minimax.io/anthropic/messages" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $API_KEY" \
      -H "anthropic-version: 2023-06-01" \
      -d '{"model":"MiniMax-M2.1","messages":[{"role":"user","content":"你好"}],"max_tokens":100}')
    
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "Response from Global endpoint:"
    echo "Status: $http_code"
    echo "Body: $response_body"
    echo ""
    
    if [ "$http_code" -eq 200 ]; then
        echo "✓ Minimax Global endpoint is working!"
    else
        echo "✗ Minimax Global endpoint failed with status $http_code"
        echo "Trying without /messages suffix on CN endpoint..."
        
        response=$(curl -s -w "\n%{http_code}" -X POST "https://api.minimaxi.com/anthropic" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $API_KEY" \
          -H "anthropic-version: 2023-06-01" \
          -d '{"model":"MiniMax-M2.1","messages":[{"role":"user","content":"你好"}],"max_tokens":100}')
        
        http_code=$(echo "$response" | tail -n1)
        response_body=$(echo "$response" | sed '$d')
        
        echo "Response from CN endpoint (without /messages):"
        echo "Status: $http_code"
        echo "Body: $response_body"
        echo ""
        
        if [ "$http_code" -eq 200 ]; then
            echo "✓ Minimax CN endpoint (without /messages) is working!"
        else
            echo "✗ All Minimax endpoints failed"
        fi
    fi
fi