package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestRoot_Help just exercises the top-level cobra wiring; ensures we
// can build and serialise the help text without panicking.
func TestRoot_Help(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --help: %v", err)
	}
	if !strings.Contains(buf.String(), "analyze") {
		t.Fatalf("help missing 'analyze' command:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), accuracyDisclosure) {
		t.Fatalf("help missing accuracy disclosure:\n%s", buf.String())
	}
}

// TestDemo_RunsAndIncludesDisclosure exercises the full demo path
// (embedded fixture → parser → rules → render).
func TestDemo_RunsAndIncludesDisclosure(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"demo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute demo: %v\n%s", err, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"Sevro Sandbox Analysis",
		"workload: api",
		"workload: worker",
		"±40%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in demo output:\n%s", want, out)
		}
	}
}

// TestAnalyze_FixtureFile exercises the analyze command against the
// versioned testdata fixture.
func TestAnalyze_FixtureFile(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"analyze", "../../testdata/fixtures/basic-chart/values.yaml"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute analyze: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "Findings (2)") {
		t.Fatalf("expected 2 findings header, got:\n%s", out)
	}
}

// TestAnalyze_JSONShape exercises --json on a fixture and validates
// the schema is intact.
func TestAnalyze_JSONShape(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"analyze", "--json", "../../testdata/fixtures/basic-chart/values.yaml"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute --json: %v\n%s", err, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		`"accuracy_disclosure"`,
		`"workloads_analyzed": 3`,
		`"DetectorID": "cpu-overprovisioned"`,
		`"DetectorID": "missing-memory-limit"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in JSON output:\n%s", want, out)
		}
	}
}
