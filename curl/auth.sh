#!/usr/bin/env bash

set -e

BASE_URL=http://localhost:8080

TOKEN=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"john@gmail.com","password":"qwe5qweqwe"}' \
    | jq -r '.token')

export TOKEN
echo "Token loaded"
