#!/bin/bash

# Set the login credentials
USERNAME="patron"
PASSWORD="password1!"
PATRON_IP="192.168.50.240"
PATRON_API_PORT="8080"

# Perform the login request
LOGIN_RESPONSE=$(curl -s -X POST http://${PATRON_IP}:${PATRON_API_PORT}/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\"}")

# Extract the token from the response
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

# Check if token is not null
if [ "$TOKEN" == "null" ]; then
  echo "Login failed: $(echo $LOGIN_RESPONSE | jq -r '.error')"
  exit 1
fi

echo "Login successful. Token: $TOKEN"

# Use the token to access a protected endpoint
RESPONSE=$(curl -s -X GET http://${PATRON_IP}:${PATRON_API_PORT}/api/data \
  -H "Authorization: $TOKEN")

echo "Response from /api/data: $RESPONSE"

# Set the new user details
NEW_USERNAME="testuser"
NEW_PASSWORD="testpassword"
NEW_ROLE="admin"

# Create a new user
CREATE_USER_RESPONSE=$(curl -s -X POST http://${PATRON_IP}:${PATRON_API_PORT}/users \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d "{\"username\":\"${NEW_USERNAME}\",\"password_hash\":\"${NEW_PASSWORD}\",\"role\":\"${NEW_ROLE}\"}")

echo "Response from creating new user: $CREATE_USER_RESPONSE"

