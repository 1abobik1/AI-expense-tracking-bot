package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

const (
	TelegramPolling = "polling" // сервер сам забирает сообщения; публичный адрес не нужен
	TelegramWebhook = "webhook" // Telegram присылает сообщения на /telegram/webhook
	TelegramOff     = "off"     // только веб-панель
)

type Config struct {
	Timezone Location `env:"APP_TIMEZONE" envDefault:"Europe/Moscow"`
	HTTPAddr string   `env:"HTTP_ADDR" envDefault:":8080"`
	DBPath   string   `env:"DB_PATH" envDefault:"data/expenses.db"` // в Worker вместо файла D1
	LLM      LLMConfig
	Telegram TelegramConfig
	Limits   LimitsConfig
	Admin    AdminConfig
}

type LLMConfig struct {
	Provider  string     `env:"LLM_PROVIDER" envDefault:"openai" validate:"oneof=openai openrouter anthropic workersai"`
	Model     string     `env:"LLM_MODEL" validate:"required"`
	BaseURL   string     `env:"LLM_BASE_URL" validate:"omitempty,url"`
	APIKey    string     `env:"LLM_API_KEY"`
	ExtraBody JSONObject `env:"LLM_EXTRA_BODY"`
}

type TelegramConfig struct {
	Mode          string `env:"TELEGRAM_MODE" envDefault:"polling" validate:"oneof=polling webhook off"`
	Token         string `env:"TELEGRAM_BOT_TOKEN" validate:"required_unless=Mode off"`
	WebhookSecret string `env:"TELEGRAM_WEBHOOK_SECRET" validate:"required_if=Mode webhook"`
	APIBase       string `env:"TELEGRAM_API_BASE" validate:"omitempty,url"`
}

type LimitsConfig struct {
	MessagesPerUserPerDay int `env:"BOT_MESSAGES_PER_USER_PER_DAY" envDefault:"100" validate:"gte=0"` // 0 — без лимита
	LLMCallsPerDay        int `env:"LLM_CALLS_PER_DAY" envDefault:"1000" validate:"gte=0"`
	MaxMessageLength      int `env:"BOT_MAX_MESSAGE_LENGTH" envDefault:"500" validate:"gte=0"`
}

type AdminConfig struct {
	Password             string `env:"ADMIN_PASSWORD" validate:"min=8"`
	LoginAttemptsPerHour int    `env:"ADMIN_LOGIN_ATTEMPTS_PER_HOUR" envDefault:"10" validate:"gte=0"`
}

// Load читает конфиг; getenv — os.Getenv или cloudflare.Getenv.
func Load(getenv func(string) string) (*Config, error) {
	var cfg Config

	environment, err := environment(&cfg, getenv)
	if err != nil {
		return nil, err
	}

	if err := env.ParseWithOptions(&cfg, env.Options{Environment: environment}); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// environment собирает значения нужных переменных: в Worker нет списка всех env, только чтение по имени.
func environment(cfg *Config, getenv func(string) string) (map[string]string, error) {
	params, err := env.GetFieldParams(cfg)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string, len(params))
	for _, p := range params {
		if v := getenv(p.Key); v != "" {
			out[p.Key] = v
		}
	}

	return out, nil
}

// validate возвращает ошибки с именами переменных окружения: «ADMIN_PASSWORD: min 8».
func validate(cfg *Config) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(f reflect.StructField) string { return f.Tag.Get("env") })
	err := v.Struct(cfg)

	var fieldErrs validator.ValidationErrors
	if !errors.As(err, &fieldErrs) {
		return err
	}

	msgs := make([]string, 0, len(fieldErrs))
	for _, fe := range fieldErrs {
		msgs = append(msgs, strings.TrimSpace(fmt.Sprintf("%s: %s %s", fe.Field(), fe.Tag(), fe.Param())))
	}

	return errors.New("invalid config: " + strings.Join(msgs, "; "))
}

// Location — часовой пояс из названия («Europe/Moscow»).
type Location struct{ *time.Location }

func (l *Location) UnmarshalText(b []byte) error {
	loc, err := time.LoadLocation(string(b))
	l.Location = loc

	return err
}

// JSONObject — параметры модели, которые уходят в запрос как есть.
type JSONObject map[string]any

func (o *JSONObject) UnmarshalText(b []byte) error {
	return json.Unmarshal(b, (*map[string]any)(o))
}
