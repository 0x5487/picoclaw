package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sipeed/picoclaw/pkg/agent/sandbox"
	"github.com/sipeed/picoclaw/pkg/fileutil"
)

type ReadFileTool struct {
	allowPaths []*regexp.Regexp
}

func NewReadFileTool(workspace string, restrict bool, allowPaths ...[]*regexp.Regexp) *ReadFileTool {
	return &ReadFileTool{allowPaths: firstPatternSet(allowPaths)}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ErrorResult("path is required")
	}

	if content, handled, err := readAllowedHostPath(ctx, path, t.allowPaths); handled {
		if err != nil {
			return ErrorResult(fmt.Sprintf("failed to read file: %v", err))
		}
		return NewToolResult(string(content))
	}

	sb := sandbox.FromContext(ctx)
	if sb == nil {
		return ErrorResult("sandbox environment unavailable")
	}

	content, err := sb.Fs().ReadFile(ctx, path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to read file: %v", err))
	}
	return NewToolResult(string(content))
}

type WriteFileTool struct {
	allowPaths []*regexp.Regexp
}

func NewWriteFileTool(workspace string, restrict bool, allowPaths ...[]*regexp.Regexp) *WriteFileTool {
	return &WriteFileTool{allowPaths: firstPatternSet(allowPaths)}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file"
}

func (t *WriteFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		return ErrorResult("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return ErrorResult("content is required")
	}

	if handled, err := writeAllowedHostPath(ctx, path, []byte(content), t.allowPaths); handled {
		if err != nil {
			return ErrorResult(fmt.Sprintf("failed to write file: %v", err))
		}
		return SilentResult(fmt.Sprintf("File written: %s", path))
	}

	sb := sandbox.FromContext(ctx)
	if sb == nil {
		return ErrorResult("sandbox environment unavailable")
	}

	if err := sb.Fs().WriteFile(ctx, path, []byte(content), true); err != nil {
		return ErrorResult(fmt.Sprintf("failed to write file: %v", err))
	}
	return SilentResult(fmt.Sprintf("File written: %s", path))
}

type ListDirTool struct {
	allowPaths []*regexp.Regexp
}

func NewListDirTool(workspace string, restrict bool, allowPaths ...[]*regexp.Regexp) *ListDirTool {
	return &ListDirTool{allowPaths: firstPatternSet(allowPaths)}
}

func (t *ListDirTool) Name() string {
	return "list_dir"
}

func (t *ListDirTool) Description() string {
	return "List files and directories in a path"
}

func (t *ListDirTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to list",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ListDirTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	path, ok := args["path"].(string)
	if !ok {
		path = "."
	}

	if entries, handled, err := readAllowedHostDir(ctx, path, t.allowPaths); handled {
		if err != nil {
			return ErrorResult(fmt.Sprintf("failed to read directory: %v", err))
		}
		return formatDirEntries(entries)
	}

	sb := sandbox.FromContext(ctx)
	if sb == nil {
		return ErrorResult("sandbox environment unavailable")
	}

	entries, err := sb.Fs().ReadDir(ctx, path)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to read directory: %v", err))
	}
	return formatDirEntries(entries)
}

func formatDirEntries(entries []os.DirEntry) *ToolResult {
	var result strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			result.WriteString("DIR:  " + entry.Name() + "\n")
		} else {
			result.WriteString("FILE: " + entry.Name() + "\n")
		}
	}
	return NewToolResult(result.String())
}

func firstPatternSet(sets [][]*regexp.Regexp) []*regexp.Regexp {
	if len(sets) == 0 {
		return nil
	}
	return sets[0]
}

func validatePath(path, workspace string, restrict bool) (string, error) {
	return sandbox.ValidatePath(path, workspace, restrict)
}

func hostSandboxFromContext(ctx context.Context) *sandbox.HostSandbox {
	sb := sandbox.FromContext(ctx)
	if sb == nil {
		return nil
	}
	host, _ := sb.(*sandbox.HostSandbox)
	return host
}

func matchesAnyPattern(path string, patterns []*regexp.Regexp) bool {
	if len(patterns) == 0 {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	for _, pattern := range patterns {
		if pattern != nil && (pattern.MatchString(path) || pattern.MatchString(absPath)) {
			return true
		}
	}
	return false
}

func readAllowedHostPath(ctx context.Context, path string, patterns []*regexp.Regexp) ([]byte, bool, error) {
	if hostSandboxFromContext(ctx) == nil || !matchesAnyPattern(path, patterns) {
		return nil, false, nil
	}
	content, err := os.ReadFile(path)
	return content, true, err
}

func writeAllowedHostPath(ctx context.Context, path string, data []byte, patterns []*regexp.Regexp) (bool, error) {
	if hostSandboxFromContext(ctx) == nil || !matchesAnyPattern(path, patterns) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return true, err
	}
	return true, fileutil.WriteFileAtomic(path, data, 0o644)
}

func readAllowedHostDir(ctx context.Context, path string, patterns []*regexp.Regexp) ([]os.DirEntry, bool, error) {
	if hostSandboxFromContext(ctx) == nil || !matchesAnyPattern(path, patterns) {
		return nil, false, nil
	}
	entries, err := os.ReadDir(path)
	return entries, true, err
}
