package mcp_test

import (
	"reflect"
	"strings"
	"testing"

	nextdoor "github.com/teslashibe/nextdoor-go"
	ndmcp "github.com/teslashibe/nextdoor-go/mcp"
	"github.com/teslashibe/mcptool"
)

// TestEveryClientMethodIsWrappedOrExcluded fails when a new exported method
// is added to *nextdoor.Client without either being wrapped by an MCP tool
// or being added to ndmcp.Excluded with a reason. This is the drift-
// prevention mechanism: keeping the MCP surface in lockstep with the package
// API is enforced by CI rather than convention.
func TestEveryClientMethodIsWrappedOrExcluded(t *testing.T) {
	rep := mcptool.Coverage(
		reflect.TypeOf(&nextdoor.Client{}),
		ndmcp.Provider{}.Tools(),
		ndmcp.Excluded,
	)
	if len(rep.Missing) > 0 {
		t.Fatalf("methods missing MCP exposure (add a tool or list in excluded.go): %v", rep.Missing)
	}
	if len(rep.UnknownExclusions) > 0 {
		t.Fatalf("excluded.go references methods that don't exist on *Client (rename?): %v", rep.UnknownExclusions)
	}
	if len(rep.Wrapped)+len(rep.Excluded) == 0 {
		t.Fatal("no wrapped or excluded methods detected — coverage helper is mis-configured")
	}
}

// TestToolsValidate verifies every tool has a non-empty name in canonical
// snake_case form, a description within length limits, and a non-nil Invoke
// + InputSchema.
func TestToolsValidate(t *testing.T) {
	if err := mcptool.ValidateTools(ndmcp.Provider{}.Tools()); err != nil {
		t.Fatal(err)
	}
}

// TestPlatformName guards against accidental rebrands.
func TestPlatformName(t *testing.T) {
	if got := (ndmcp.Provider{}).Platform(); got != "nextdoor" {
		t.Errorf("Platform() = %q, want nextdoor", got)
	}
}

// TestToolsHaveNextdoorPrefix encodes the per-platform naming convention.
func TestToolsHaveNextdoorPrefix(t *testing.T) {
	for _, tool := range (ndmcp.Provider{}).Tools() {
		if !strings.HasPrefix(tool.Name, "nextdoor_") {
			t.Errorf("tool %q lacks nextdoor_ prefix", tool.Name)
		}
	}
}
