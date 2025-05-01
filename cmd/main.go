package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agl/balancer/balancer"
	"github.com/agl/balancer/config"
	"github.com/agl/balancer/ratelimit"
)

func main() {
	cfg, err := loadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	lb, err := balancer.NewLoadBalancer(cfg)
	if err != nil {
		log.Fatalf("Error creating load balancer: %v", err)
	}
	defer lb.Stop()

	var handler http.Handler = lb

	if cfg.RateLimit.Enabled {
		rl := ratelimit.NewRateLimiter()
		defer rl.Stop()
		handler = ratelimit.Middleware(rl)(lb)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")
}

func loadConfig(path string) (config.Config, error) {
    var cfg config.Config
    
    file, err := os.ReadFile(path)
    if err != nil {
        return cfg, err
    }

    if err := json.Unmarshal(file, &cfg); err != nil {
        return cfg, err
    }

    if cfg.Port == "" {
        cfg.Port = "8080"
    }
    if cfg.HealthCheck == 0 {
        cfg.HealthCheck = 10 * time.Second
    }

    return cfg, nil
}