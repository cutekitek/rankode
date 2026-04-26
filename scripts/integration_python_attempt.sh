#!/usr/bin/env bash
set -euo pipefail

BACKEND_CONTAINER="${BACKEND_CONTAINER:-rankode_backend}"
API_PREFIX="${API_PREFIX:-/api}"
POLL_TIMEOUT_SECONDS="${POLL_TIMEOUT_SECONDS:-60}"
POLL_INTERVAL_SECONDS="${POLL_INTERVAL_SECONDS:-2}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

discover_base_url() {
  if [[ -n "${BASE_URL:-}" ]]; then
    printf '%s\n' "${BASE_URL%/}"
    return
  fi

  local published
  published="$(docker port "$BACKEND_CONTAINER" 4000/tcp 2>/dev/null | head -n 1 || true)"
  if [[ -z "$published" ]]; then
    echo "cannot discover backend port from Docker container '$BACKEND_CONTAINER'; set BASE_URL=http://host:port" >&2
    exit 1
  fi

  local port
  port="${published##*:}"
  if [[ -z "$port" || "$port" == "$published" ]]; then
    echo "cannot parse backend port from docker output: $published" >&2
    exit 1
  fi

  printf 'http://localhost:%s\n' "$port"
}

request_json() {
  local expected_status="$1"
  local method="$2"
  local url="$3"
  local payload="$4"
  local token="${5:-}"
  local body_file status

  body_file="$(mktemp)"
  if [[ -n "$token" ]]; then
    status="$(curl -sS -o "$body_file" -w '%{http_code}' \
      -X "$method" "$url" \
      -H 'Content-Type: application/json' \
      -H "Authorization: Bearer $token" \
      --data "$payload")"
  else
    status="$(curl -sS -o "$body_file" -w '%{http_code}' \
      -X "$method" "$url" \
      -H 'Content-Type: application/json' \
      --data "$payload")"
  fi

  if [[ "$status" != "$expected_status" ]]; then
    echo "request failed: $method $url" >&2
    echo "expected HTTP $expected_status, got HTTP $status" >&2
    echo "response body:" >&2
    cat "$body_file" >&2
    rm -f "$body_file"
    exit 1
  fi

  cat "$body_file"
  rm -f "$body_file"
}

request_file_upload() {
  local expected_status="$1"
  local url="$2"
  local token="$3"
  local file_path="$4"
  local body_file status

  body_file="$(mktemp)"
  status="$(curl -sS -o "$body_file" -w '%{http_code}' \
    -X POST "$url" \
    -H "Authorization: Bearer $token" \
    -F "file=@${file_path};type=text/plain")"

  if [[ "$status" != "$expected_status" ]]; then
    echo "upload failed: POST $url" >&2
    echo "expected HTTP $expected_status, got HTTP $status" >&2
    echo "response body:" >&2
    cat "$body_file" >&2
    rm -f "$body_file"
    exit 1
  fi

  rm -f "$body_file"
}

require_cmd curl
require_cmd docker
require_cmd jq

BASE_URL="$(discover_base_url)"
API_BASE="${BASE_URL}${API_PREFIX}"
RUN_ID="$(date +%s%N)"
PASSWORD="${PASSWORD:-Password123!}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

printf '2 3\n' > "$tmpdir/input.txt"
printf '5\n' > "$tmpdir/output.txt"

teacher_user="teacher_${RUN_ID}"
teacher_email="${teacher_user}@example.test"
student_user="student_${RUN_ID}"
student_email="${student_user}@example.test"

echo "Using API: $API_BASE"
echo "Registering teacher: $teacher_user"
teacher_register_payload="$(jq -n \
  --arg username "$teacher_user" \
  --arg email "$teacher_email" \
  --arg password "$PASSWORD" \
  '{username: $username, email: $email, password: $password}')"
teacher="$(request_json 201 POST "$API_BASE/auth/register" "$teacher_register_payload")"
teacher_id="$(jq -r '.id' <<<"$teacher")"

teacher_login_payload="$(jq -n \
  --arg identifier "$teacher_user" \
  --arg password "$PASSWORD" \
  '{identifier: $identifier, password: $password}')"
teacher_login="$(request_json 200 POST "$API_BASE/auth/login" "$teacher_login_payload")"
teacher_token="$(jq -r '.token' <<<"$teacher_login")"

echo "Creating Python task as teacher ID $teacher_id"
task_payload="$(jq -n \
  --arg title "Python sum ${RUN_ID}" \
  '{title: $title, description: "Read two integers from stdin and print their sum.", difficulty: 1, topics: []}')"
task="$(request_json 201 POST "$API_BASE/tasks" "$task_payload" "$teacher_token")"
task_id="$(jq -r '.id' <<<"$task")"

echo "Creating and uploading test case for task ID $task_id"
test_case_payload="$(jq -n --argjson task_id "$task_id" '{task_id: $task_id}')"
test_case="$(request_json 201 POST "$API_BASE/test-cases" "$test_case_payload" "$teacher_token")"
test_case_id="$(jq -r '.id' <<<"$test_case")"
request_file_upload 200 "$API_BASE/test-cases/$test_case_id/upload?type=input" "$teacher_token" "$tmpdir/input.txt"
request_file_upload 200 "$API_BASE/test-cases/$test_case_id/upload?type=output" "$teacher_token" "$tmpdir/output.txt"

echo "Registering student: $student_user"
student_register_payload="$(jq -n \
  --arg username "$student_user" \
  --arg email "$student_email" \
  --arg password "$PASSWORD" \
  '{username: $username, email: $email, password: $password}')"
student="$(request_json 201 POST "$API_BASE/auth/register" "$student_register_payload")"
student_id="$(jq -r '.id' <<<"$student")"

student_login_payload="$(jq -n \
  --arg identifier "$student_user" \
  --arg password "$PASSWORD" \
  '{identifier: $identifier, password: $password}')"
student_login="$(request_json 200 POST "$API_BASE/auth/login" "$student_login_payload")"
student_token="$(jq -r '.token' <<<"$student_login")"

student_code='import sys

nums = list(map(int, sys.stdin.read().split()))
print(sum(nums))
'

echo "Submitting Python attempt as student ID $student_id"
attempt_payload="$(jq -n \
  --argjson task_id "$task_id" \
  --arg code "$student_code" \
  '{task_id: $task_id, lang: "python3", code: $code}')"
request_json 201 POST "$API_BASE/attempts" "$attempt_payload" "$student_token" >/dev/null

echo "Waiting for attempt completion"
deadline=$((SECONDS + POLL_TIMEOUT_SECONDS))
latest_attempt='[]'
latest_status=''
while (( SECONDS < deadline )); do
  latest_attempt="$(curl -sS \
    -H "Authorization: Bearer $student_token" \
    "$API_BASE/attempts?taskId=$task_id")"

  latest_status="$(jq -r '.[0].status // empty' <<<"$latest_attempt")"
  if [[ -n "$latest_status" && "$latest_status" != "4" ]]; then
    break
  fi

  sleep "$POLL_INTERVAL_SECONDS"
done

if [[ -z "$latest_status" ]]; then
  echo "no attempt was returned by GET /attempts?taskId=$task_id" >&2
  echo "$latest_attempt" >&2
  exit 1
fi

if [[ "$latest_status" == "4" ]]; then
  echo "attempt did not complete within ${POLL_TIMEOUT_SECONDS}s" >&2
  echo "$latest_attempt" | jq . >&2
  exit 1
fi

if [[ "$latest_status" != "0" ]]; then
  echo "attempt completed with non-success status: $latest_status" >&2
  echo "$latest_attempt" | jq . >&2
  exit 1
fi

attempt_id="$(jq -r '.[0].id' <<<"$latest_attempt")"
running_time="$(jq -r '.[0].running_time' <<<"$latest_attempt")"
memory_usage="$(jq -r '.[0].memory_usage' <<<"$latest_attempt")"

echo "Integration test passed"
echo "teacher_id=$teacher_id"
echo "student_id=$student_id"
echo "task_id=$task_id"
echo "test_case_id=$test_case_id"
echo "attempt_id=$attempt_id"
echo "status=$latest_status"
echo "running_time=$running_time"
echo "memory_usage=$memory_usage"
