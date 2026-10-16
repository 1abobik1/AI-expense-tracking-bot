package config

import (
	"strings"
	"testing"
)

func load(env map[string]string) (*Config, error) {
	return Load(func(k string) string { return env[k] })
}

func valid() map[string]string {
	return map[string]string{"ADMIN_PASSWORD": "password1", "LLM_MODEL": "m", "TELEGRAM_BOT_TOKEN": "t"}
}

func TestDefaultsAndParsing(t *testing.T) {
	env := valid()
	env["LLM_EXTRA_BODY"] = `{"chat_template_kwargs":{"enable_thinking":false}}`
	env["APP_TIMEZONE"] = "Asia/Novosibirsk"

	cfg, err := load(env)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Telegram.Mode != TelegramPolling || cfg.Limits.MessagesPerUserPerDay != 100 || cfg.DBPath != "data/expenses.db" {
		t.Errorf("defaults: %+v", cfg)
	}

	if cfg.Timezone.String() != "Asia/Novosibirsk" || cfg.LLM.ExtraBody["chat_template_kwargs"] == nil {
		t.Errorf("parsed: tz=%v extra=%v", cfg.Timezone, cfg.LLM.ExtraBody)
	}
}

func TestValidation(t *testing.T) {
	cases := map[string]struct {
		set  map[string]string
		want string
	}{
		"short password":     {map[string]string{"ADMIN_PASSWORD": "short"}, "ADMIN_PASSWORD: min 8"},
		"webhook w/o secret": {map[string]string{"TELEGRAM_MODE": "webhook"}, "TELEGRAM_WEBHOOK_SECRET: required_if"},
		"unknown provider":   {map[string]string{"LLM_PROVIDER": "gpt"}, "LLM_PROVIDER: oneof"},
		"bad timezone":       {map[string]string{"APP_TIMEZONE": "Mars/Base"}, `"Timezone"`},
		"bad extra body":     {map[string]string{"LLM_EXTRA_BODY": "{"}, `"ExtraBody"`},
		"negative limit":     {map[string]string{"LLM_CALLS_PER_DAY": "-1"}, "LLM_CALLS_PER_DAY: gte 0"},
	}
	for name, c := range cases {
		env := valid()
		for k, v := range c.set {
			env[k] = v
		}

		if _, err := load(env); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want %q", name, err, c.want)
		}
	}
}

func TestTokenNotRequiredWhenBotOff(t *testing.T) {
	if _, err := load(map[string]string{"ADMIN_PASSWORD": "password1", "LLM_MODEL": "m", "TELEGRAM_MODE": "off"}); err != nil {
		t.Fatal(err)
	}

	if _, err := load(map[string]string{"ADMIN_PASSWORD": "password1", "LLM_MODEL": "m"}); err == nil {
		t.Fatal("token must be required in polling mode")
	}
}
