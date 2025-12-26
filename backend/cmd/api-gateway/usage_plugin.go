package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	usageevent "github.com/your-org/api-usage-billing/backend/internal/event/usage"
)

// Config defines plugin configuration for usage logging.
type Config struct {
	RedisHost     string `json:"redis_host"`
	RedisPort     int    `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
	StreamPrefix  string `json:"stream_prefix"`
	Stream        string `json:"stream"`
	MaxLen        int64  `json:"max_len"`

	initOnce  sync.Once
	initErr   error
	publisher *usageevent.Publisher
}

// New returns a new plugin configuration instance.
func New() interface{} {
	return &Config{
		RedisHost:    "redis.default.svc.cluster.local",
		RedisPort:    6379,
		StreamPrefix: "usage",
		Stream:       "requests",
		MaxLen:       100000,
	}
}

// Log executes in the Kong log phase after the response is sent.
func (conf *Config) Log(kong *pdk.PDK) {
	if err := conf.ensurePublisher(); err != nil {
		_ = kong.Log.Err(err.Error())
		return
	}

	orgID := headerUUID(kong, "X-Organization-ID")
	customerID := headerUUID(kong, "X-Customer-ID")
	if orgID == uuid.Nil || customerID == uuid.Nil {
		return
	}

	var apiKeyID *uuid.UUID
	if parsed := headerUUID(kong, "X-API-Key-ID"); parsed != uuid.Nil {
		apiKeyID = &parsed
	}

	requestID, _ := kong.Request.GetHeader("X-Request-ID")
	endpoint, _ := kong.Request.GetPath()
	method, _ := kong.Request.GetMethod()
	statusCode, _ := kong.Response.GetStatus()
	userAgent, _ := kong.Request.GetHeader("User-Agent")
	clientIP, _ := kong.Request.GetHeader("X-Forwarded-For")
	if clientIP == "" {
		clientIP, _ = kong.Request.GetHeader("X-Real-IP")
	}

	responseSize := parseIntHeader(kong, "Content-Length")
	proxyLatency := parseIntHeader(kong, "X-Kong-Proxy-Latency")
	upstreamLatency := parseIntHeader(kong, "X-Kong-Upstream-Latency")
	latencyMs := proxyLatency + upstreamLatency

	event := &usageevent.Event{
		ID:                uuid.NewString(),
		OrganizationID:    orgID,
		CustomerID:        customerID,
		APIKeyID:          apiKeyID,
		RequestID:         requestID,
		Endpoint:          endpoint,
		Method:            method,
		StatusCode:        statusCode,
		RequestSizeBytes:  0,
		ResponseSizeBytes: responseSize,
		LatencyMs:         latencyMs,
		RecordedAt:        time.Now().UTC(),
		UserAgent:         userAgent,
		ClientIP:          clientIP,
		Metadata:          map[string]interface{}{},
	}

	if _, err := conf.publisher.Publish(context.Background(), event); err != nil {
		_ = kong.Log.Err(err.Error())
	}
}

func (conf *Config) ensurePublisher() error {
	conf.initOnce.Do(func() {
		addr := fmt.Sprintf("%s:%d", conf.RedisHost, conf.RedisPort)
		rdb := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: conf.RedisPassword,
		})

		conf.publisher = usageevent.NewPublisher(rdb, usageevent.PublisherConfig{
			StreamPrefix: conf.StreamPrefix,
			Stream:       conf.Stream,
			MaxLen:       conf.MaxLen,
		})
	})

	return conf.initErr
}

func headerUUID(kong *pdk.PDK, name string) uuid.UUID {
	value, _ := kong.Request.GetHeader(name)
	if value == "" {
		return uuid.Nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}

func parseIntHeader(kong *pdk.PDK, name string) int {
	value, _ := kong.Request.GetHeader(name)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func main() {
	server.StartServer(New, "0.1", 0)
}
