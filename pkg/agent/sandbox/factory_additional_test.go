package sandbox

import (
	"context"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

func TestExpandHomePath(t *testing.T) {
	if got := expandHomePath(""); got != "" {
		t.Fatalf("expandHomePath(\"\") = %q, want empty", got)
	}
	if got := expandHomePath("abc"); got != "abc" {
		t.Fatalf("expandHomePath(abc) = %q", got)
	}
	if got := expandHomePath("~"); got == "" {
		t.Fatal("expandHomePath(~) should resolve to home")
	}
	if got := expandHomePath("~/x"); got == "" || got == "~/x" {
		t.Fatalf("expandHomePath(~/x) = %q, expected resolved path", got)
	}
}

func TestNewFromConfig_HostMode(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Sandbox.Mode = "off"

	sb := NewFromConfig(t.TempDir(), true, cfg)
	if _, ok := sb.(*HostSandbox); !ok {
		t.Fatalf("expected HostSandbox, got %T", sb)
	}
	if err := sb.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}

func TestNewFromConfig_AllModeReturnsUnavailableWhenBlocked(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Sandbox.Mode = "all"
	cfg.Agents.Defaults.Sandbox.Docker.Network = "host"
	cfg.Agents.Defaults.Sandbox.Prune.IdleHours = 0
	cfg.Agents.Defaults.Sandbox.Prune.MaxAgeDays = 0

	sb := NewFromConfig(t.TempDir(), true, cfg)
	if _, ok := sb.(*unavailableSandbox); !ok {
		t.Fatalf("expected unavailableSandbox, got %T", sb)
	}
	if err := sb.Start(context.Background()); err == nil {
		t.Fatal("expected unavailable sandbox start error")
	}
}
