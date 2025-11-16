#!/bin/bash

API_URL="http://localhost:8080"

# Создаём команду backend
curl -X POST "$API_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "550e8400-e29b-41d4-a716-446655440001", "username": "Alice", "is_active": true},
      {"user_id": "550e8400-e29b-41d4-a716-446655440002", "username": "Bob", "is_active": true},
      {"user_id": "550e8400-e29b-41d4-a716-446655440003", "username": "Charlie", "is_active": true},
      {"user_id": "550e8400-e29b-41d4-a716-446655440004", "username": "Dave", "is_active": true},
      {"user_id": "550e8400-e29b-41d4-a716-446655440005", "username": "Eve", "is_active": true}
    ]
  }'

# Создаём команду frontend
curl -X POST "$API_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "frontend",
    "members": [
      {"user_id": "550e8400-e29b-41d4-a716-446655440006", "username": "Frank", "is_active": true},
      {"user_id": "550e8400-e29b-41d4-a716-446655440007", "username": "Grace", "is_active": true},
      {"user_id": "550e8400-e29b-41d4-a716-446655440008", "username": "Henry", "is_active": true}
    ]
  }'

echo "Seed data created successfully"
