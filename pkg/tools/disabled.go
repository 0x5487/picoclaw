package tools

import (
	"context"
)

// DisabledTool keeps a stable tool surface but always returns an error result.
type DisabledTool struct {
	name        string
	description string
	reason      string
}

func NewDisabledTool(name, description, reason string) *DisabledTool {
	return &DisabledTool{
		name:        name,
		description: description,
		reason:      reason,
	}
}

func (t *DisabledTool) Name() string {
	return t.name
}

func (t *DisabledTool) Description() string {
	if t.description != "" {
		return t.description
	}
	return "This tool is disabled in current sandbox policy."
}

func (t *DisabledTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *DisabledTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	msg := t.reason
	if msg == "" {
		msg = "tool is disabled"
	}
	return ErrorResult(msg)
}
