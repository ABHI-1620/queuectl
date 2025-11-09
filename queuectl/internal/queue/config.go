package queue

import (
	"database/sql"
	"fmt"
)

type Config struct {
	MaxRetries     int
	BackoffBase    int
	LeaseSeconds   int
	JobTimeout     int
	PollIntervalMS int
}

func LoadConfig(db *sql.DB) Config {
	cfg := Config{
		MaxRetries:     3,
		BackoffBase:    2,
		LeaseSeconds:   60,
		JobTimeout:     30,
		PollIntervalMS: 500,
	}

	rows, err := db.Query(`SELECT key, value FROM config`)
	if err != nil {
		return cfg
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		switch k {
		case "max_retries":
			fmt.Sscanf(v, "%d", &cfg.MaxRetries)
		case "backoff_base":
			fmt.Sscanf(v, "%d", &cfg.BackoffBase)
		case "lease_seconds":
			fmt.Sscanf(v, "%d", &cfg.LeaseSeconds)
		case "job_timeout_seconds":
			fmt.Sscanf(v, "%d", &cfg.JobTimeout)
		case "poll_interval_ms":
			fmt.Sscanf(v, "%d", &cfg.PollIntervalMS)
		}
	}

	return cfg
}
