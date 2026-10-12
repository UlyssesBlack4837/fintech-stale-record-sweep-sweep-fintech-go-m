package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/infrai-examples/fintech-stale-record-sweep/internal/cleanup"
	"github.com/infrai-examples/fintech-stale-record-sweep/internal/infrai"
)

type sweepRequest struct {
	Events []cleanup.PaymentEvent `json:"events"`
}

type sweepResponse struct {
	Results []cleanup.Result `json:"results"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: sweep-service serve|schedule")
	}

	switch os.Args[1] {
	case "serve":
		serve()
	case "schedule":
		if err := schedule(context.Background()); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func serve() {
	policy := cleanup.Policy{StaleAfter: 90 * 24 * time.Hour, RiskFloor: 70}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sweep", func(w http.ResponseWriter, r *http.Request) {
		var input sweepRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid sweep input", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sweepResponse{Results: policy.Sweep(input.Events, time.Now().UTC())}); err != nil {
			log.Printf("encode sweep response: %v", err)
		}
	})

	addr := envOr("LISTEN_ADDR", ":8080")
	log.Printf("sweep endpoint listening on %s", addr)
	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}

func schedule(ctx context.Context) error {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return errors.New("INFRAI_API_KEY is required")
	}
	taskURL := os.Getenv("SWEEP_TASK_URL")
	if taskURL == "" {
		return errors.New("SWEEP_TASK_URL is required")
	}

	client := infrai.New(key)
	job, err := client.CreateCron(ctx, infrai.CreateCronRequest{
		CronExpr: envOr("SWEEP_CRON_EXPR", "15 2 * * *"),
		Task:     taskURL,
	}, "fintech-stale-record-sweep-v1")
	if err != nil {
		var apiErr *infrai.APIError
		if errors.As(err, &apiErr) {
			return fmt.Errorf("schedule rejected (%s, HTTP %d): %s", apiErr.Code, apiErr.HTTPStatus, apiErr.Message)
		}
		return err
	}

	fmt.Printf("scheduled payment cleanup: %s\n", job.JobID)
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
