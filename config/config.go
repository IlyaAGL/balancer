package config

import "time"

type Config struct {
    Port            string        `json:"port"`
    HealthCheck     time.Duration `json:"healthCheckInterval"`
    RateLimit       RateLimitConfig `json:"rateLimit"`
    BackendServers  []ServerConfig `json:"backends"`
}

type ServerConfig struct {
    URL    string `json:"url"`
    Weight int    `json:"weight"`
    Name   string `json:"name"`
}

type RateLimitConfig struct {
    Enabled      bool `json:"enabled"`
    DefaultCap   int  `json:"defaultCapacity"`
    DefaultRate  int  `json:"defaultRate"`
}