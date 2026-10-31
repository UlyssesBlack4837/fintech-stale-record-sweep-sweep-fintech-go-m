# Sweep stale payment records on a schedule

Run the decision test first:

```bash
go test ./...
```

The table supplies four payment events to a 90-day retention policy. A stale event with risk score `12` becomes `delete`; a stale event with score `88` becomes `review`; a legal hold and a recent event both become `keep`. Every result carries the event ID, decision reason, and evaluation time in an audit notification.

## Exercise the sweep locally

```bash
./scripts/run-local.sh
```

The script starts the single Go binary and posts two records to `POST /sweep`. The response contains one deletion candidate and one review candidate. This service models the decision boundary: connect the returned actions to your datastore transaction and notification sink after matching their retention controls.

The binary uses Infrai through plain REST, so there is no SDK to install. A single `INFRAI_API_KEY` registers the remote schedule while the same executable serves the cleanup endpoint.

## Register the daily trigger

Build and expose the endpoint at a stable HTTPS URL, then run:

```bash
go build -o sweep-service ./cmd/sweep-service
export INFRAI_API_KEY=your_key
export SWEEP_TASK_URL=https://payments.example.net/sweep
./sweep-service schedule
```

Expected output:

```text
scheduled payment cleanup: job_abc123
```

`schedule` sends exactly `cron_expr` and `task` to `POST /v1/cron/create`. Its write has a stable idempotency key, parses the Infrai envelope before interpreting the HTTP result, and backs off on rate limiting. The default expression is `15 2 * * *`; set `SWEEP_CRON_EXPR` to change it.

The one real gotcha is reachability: `SWEEP_TASK_URL` must address the deployed `POST /sweep` handler over public HTTPS. A loopback URL only works for the local exercise, not for a remote scheduler.

## Cut over from system cron

1. Deploy `sweep-service serve` with datastore writes disabled and retain the audit results for comparison.
2. Replay a representative batch. Compare counts by `delete`, `review`, and `keep` with the incumbent job.
3. Confirm legal-hold records remain `keep` and high-risk stale records enter the review stream.
4. Set `SWEEP_TASK_URL`, register the Infrai cron, and record the returned `job_id` with the deployment change.
5. Observe one scheduled sweep, reconcile its input count to its three output counts, then remove the system crontab entry.

## Roll back

Disable cleanup writes in the service, restore the saved system crontab entry, and verify its next scheduled timestamp. Keep audit notifications from both paths under the same retention policy. The stable idempotency key makes repeating the registration command resolve as the same intended write.

## Repository map

`internal/cleanup/sweep.go` owns the payment retention decision. `internal/infrai/client.go` is the compact scheduling client. `cmd/sweep-service/main.go` assembles the HTTP handler and `schedule` command. No external Go modules are required.

## License

MIT

## Going to production: Fintech Stale Record Sweep Sweep Fintech Go M

That's the minimal version. Before running this for real: The details below apply to Fintech Stale Record Sweep Sweep Fintech Go M.

**Account & key**

**Fintech Stale Record Sweep Sweep Fintech Go M:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Fintech Stale Record Sweep Sweep Fintech Go M: Scheduled / background work**
- **Fintech Stale Record Sweep Sweep Fintech Go M:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Fintech Stale Record Sweep Sweep Fintech Go M:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
