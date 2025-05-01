package main

import (
	"encoding/json"
	"flag"
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
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	path := *configPath
	if path == "" {
		if envPath := os.Getenv("BACKEND_CONFIG"); envPath != "" {
			path = envPath
		} else {
			path = "config/config.json"
		}
	}

	cfg, err := loadConfig(path)
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