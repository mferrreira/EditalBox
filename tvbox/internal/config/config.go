package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Environment          string
	HTTPAddr             string
	DBPath               string
	SyncInterval         time.Duration
	TelegramToken        string
	AllowedChatIDs       map[int64]struct{}
	AgentBaseURL         string
	AgentTimeout         time.Duration
	SessionRetention     time.Duration
	TelegramPollTimeout  time.Duration
	TelegramRetryLimit   int
	TelegramRetryBackoff time.Duration
	CollectorUserAgent   string
	CollectorSeedURLs    []string
	CollectorFollowRules []string
}

func Load() Config {
	return Config{
		Environment:          env("EDITALBOX_ENV", "development"),
		HTTPAddr:             env("EDITALBOX_HTTP_ADDR", ":8080"),
		DBPath:               env("EDITALBOX_DB_PATH", "./data/editalbox.db"),
		SyncInterval:         envDuration("EDITALBOX_SYNC_INTERVAL", 24*time.Hour),
		TelegramToken:        env("EDITALBOX_TELEGRAM_TOKEN", ""),
		AllowedChatIDs:       parseChatIDs(env("EDITALBOX_TELEGRAM_ALLOWED_CHAT_IDS", "")),
		AgentBaseURL:         env("EDITALBOX_AGENT_BASE_URL", "http://127.0.0.1:8090"),
		AgentTimeout:         envDuration("EDITALBOX_AGENT_TIMEOUT", 20*time.Second),
		SessionRetention:     24 * time.Hour,
		TelegramPollTimeout:  envDuration("EDITALBOX_TELEGRAM_POLL_TIMEOUT", 25*time.Second),
		TelegramRetryLimit:   envInt("EDITALBOX_TELEGRAM_RETRY_LIMIT", 3),
		TelegramRetryBackoff: envDuration("EDITALBOX_TELEGRAM_RETRY_BACKOFF", 2*time.Second),
		CollectorUserAgent: "Mozilla/5.0 (X11; Linux armv7l) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36",
		CollectorSeedURLs: []string{
			"https://www.ifnmg.edu.br/januaria",
			"https://www.ifnmg.edu.br/editais-ifnmg",
			"https://www.ifnmg.edu.br/mais-noticias-januaria",
			"https://www.ifnmg.edu.br/assistenciaestudantil-januaria/editais-assistenciaestudantil-januaria",
			"https://www.ifnmg.edu.br/extensao-januaria/editais",
			"https://www.ifnmg.edu.br/pesquisa-januaria/pesquisa/editais-pesquisa-januaria",
			"https://www.ifnmg.edu.br/processoseletivo",
			"https://www.ifnmg.edu.br/professor-substituto",
		},
		CollectorFollowRules: []string{
			"/mais-noticias-januaria/",
			"/processoseletivo",
			"/professor-substituto/55-portal/januaria/",
			"/assistenciaestudantil-januaria/",
			"/extensao-januaria/",
			"/pesquisa-januaria/",
			"documento.ifnmg.edu.br/action.php",
		},
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	var out int
	if _, err := fmt.Sscan(value, &out); err != nil {
		return fallback
	}
	return out
}

func parseChatIDs(value string) map[int64]struct{} {
	out := map[int64]struct{}{}
	for part := range strings.SplitSeq(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var id int64
		if _, err := fmt.Sscan(part, &id); err == nil {
			out[id] = struct{}{}
		}
	}
	return out
}
