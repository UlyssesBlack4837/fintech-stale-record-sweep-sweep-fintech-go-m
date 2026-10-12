#!/bin/sh
set -eu

go run ./cmd/sweep-service serve &
service_pid=$!
trap 'kill "$service_pid"' EXIT INT TERM
sleep 1

curl --fail-with-body --request POST http://localhost:8080/sweep \
  --header 'Content-Type: application/json' \
  --data '{"events":[{"id":"pay_101","state":"settled","last_updated":"2025-01-01T00:00:00Z","risk_score":12,"legal_hold":false,"amount_cents":4200},{"id":"pay_102","state":"disputed","last_updated":"2025-01-01T00:00:00Z","risk_score":88,"legal_hold":false,"amount_cents":9900}]}'
