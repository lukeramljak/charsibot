#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8081}"
PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); printf "  \033[32m✓\033[0m %s\n" "$1"; }
fail() { FAIL=$((FAIL + 1)); printf "  \033[31m✗\033[0m %s\n" "$1"; }

printf "\nValidating cutover against %s\n\n" "$BASE_URL"

# --- Health ---
printf "Health\n"
health_status=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/health")
if [ "$health_status" = "200" ]; then
  pass "/health returns 200"
else
  fail "/health returned $health_status (expected 200)"
fi

# --- Readiness ---
printf "\nReadiness\n"
ready_response=$(curl -s "$BASE_URL/ready")
ready_status=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/ready")

if [ "$ready_status" = "200" ]; then
  pass "/ready returns 200"
else
  fail "/ready returned $ready_status (expected 200)"
fi

ready_flag=$(echo "$ready_response" | python3 -c "import sys,json; print(json.load(sys.stdin).get('ready',''))" 2>/dev/null || echo "")
if [ "$ready_flag" = "True" ] || [ "$ready_flag" = "true" ]; then
  pass "readiness reports ready=true"
else
  fail "readiness reports ready=$ready_flag (expected true)"
fi

for component in catalog database; do
  comp_ready=$(echo "$ready_response" | python3 -c "import sys,json; print(json.load(sys.stdin).get('components',{}).get('$component',''))" 2>/dev/null || echo "")
  if [ "$comp_ready" = "True" ] || [ "$comp_ready" = "true" ]; then
    pass "component $component is ready"
  else
    fail "component $component reports $comp_ready"
  fi
done

# --- Admin panel ---
printf "\nAdmin\n"
admin_status=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/admin")
if [ "$admin_status" = "200" ]; then
  pass "/admin returns 200"
else
  fail "/admin returned $admin_status (expected 200)"
fi

# --- SSE heartbeat ---
printf "\nSSE\n"
sse_output=$(timeout 3 curl -s -N "$BASE_URL/events" 2>/dev/null || true)
if echo "$sse_output" | grep -q ': ping'; then
  pass "/events sends initial heartbeat"
else
  fail "/events did not send a heartbeat within 3 seconds"
fi

# --- Database validation (requires local access) ---
printf "\nDatabase (via readiness)\n"
if [ "$ready_status" = "200" ]; then
  pass "database is accessible and schema validated at startup"
else
  fail "database readiness could not be confirmed"
fi

# --- Summary ---
TOTAL=$((PASS + FAIL))
printf "\n%d/%d checks passed" "$PASS" "$TOTAL"
if [ "$FAIL" -gt 0 ]; then
  printf " (\033[31m%d failed\033[0m)" "$FAIL"
  printf "\n"
  exit 1
else
  printf "\n"
fi
