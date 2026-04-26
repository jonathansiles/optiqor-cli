package analyze

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_ReturnsFindings(t *testing.T) {
	in := strings.NewReader(`
api:
  resources:
    requests:
      cpu: "2"
      memory: "1Gi"
    limits:
      cpu: "2.5"
      memory: "2Gi"
worker:
  resources:
    requests:
      memory: "1Gi"
`)
	rep, err := Run(in, Options{Source: "test"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if rep.Workloads != 2 {
		t.Errorf("workloads = %d, want 2", rep.Workloads)
	}
	if len(rep.Findings) == 0 {
		t.Fatal("expected findings, got 0")
	}

	var sawCPU, sawMissingLimit bool
	for _, f := range rep.Findings {
		switch f.DetectorID {
		case "cpu-overprovisioned":
			sawCPU = true
		case "missing-memory-limit":
			sawMissingLimit = true
		}
	}
	if !sawCPU {
		t.Error("missing cpu-overprovisioned finding")
	}
	if !sawMissingLimit {
		t.Error("missing missing-memory-limit finding")
	}
}

func TestRun_BadYAML(t *testing.T) {
	if _, err := Run(strings.NewReader("not: valid: yaml::"), Options{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunPath_File(t *testing.T) {
	dir := t.TempDir()
	values := filepath.Join(dir, "values.yaml")
	if err := os.WriteFile(values, []byte(`
api:
  resources:
    requests:
      cpu: "1"
      memory: "1Gi"
    limits:
      cpu: "1"
      memory: "1Gi"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := RunPath(values)
	if err != nil {
		t.Fatalf("RunPath: %v", err)
	}
	if rep.Workloads != 1 {
		t.Errorf("workloads = %d, want 1", rep.Workloads)
	}
}

func TestRunPath_Directory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(`api: {resources: {requests: {cpu: 1, memory: 1Gi}, limits: {cpu: 1, memory: 1Gi}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	rep, err := RunPath(dir)
	if err != nil {
		t.Fatalf("RunPath dir: %v", err)
	}
	if rep.Workloads != 1 {
		t.Errorf("workloads = %d", rep.Workloads)
	}
	if !strings.HasSuffix(rep.Source, "values.yaml") {
		t.Errorf("source should point to values.yaml, got %q", rep.Source)
	}
}

func TestRunPath_Missing(t *testing.T) {
	if _, err := RunPath("/nonexistent/path/values.yaml"); err == nil {
		t.Fatal("expected error")
	}
}
