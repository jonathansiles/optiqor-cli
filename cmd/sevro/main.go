// Command sevro is the entrypoint for the open-source Sevro CLI.
//
// The CLI is a deterministic rule engine that analyzes Helm charts for cost
// inefficiencies and security findings. It does NOT call any LLM and does NOT
// phone home by default — see ../../CLAUDE.md for the hard rules.
package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lowplane/sevro/internal/analyze"
	"github.com/lowplane/sevro/internal/render"
)

var version = "dev"

const accuracyDisclosure = "Sandbox accuracy: ±40%. Install the Sevro agent for exact numbers (sevro.dev/get)."

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "sevro",
		Short:         "Cost & security analysis for Kubernetes Helm charts",
		Long:          "sevro analyzes Helm charts (or values files) for cost inefficiencies and security findings.\n\n" + accuracyDisclosure,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.AddCommand(
		newAnalyzeCmd(),
		newDemoCmd(),
		newDiffCmd(),
		newScoreCmd(),
		newAuditCmd(),
		newWatchCmd(),
		newCompareCmd(),
	)

	return root
}

func newAnalyzeCmd() *cobra.Command {
	var (
		jsonOut bool
		offline bool
		share   bool
		roast   bool
	)
	cmd := &cobra.Command{
		Use:   "analyze [chart]",
		Short: "Analyze a Helm chart or values file for cost & security findings",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			rep, err := analyze.RunPath(abs)
			if err != nil {
				return err
			}
			if jsonOut {
				return render.JSON(cmd.OutOrStdout(), rep)
			}
			// `--share` and `--roast` are accepted now so flags land in
			// muscle memory; their behavior arrives in later phases.
			_ = offline
			_ = share
			_ = roast
			return render.Text(cmd.OutOrStdout(), rep)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	cmd.Flags().BoolVar(&offline, "offline", true, "do not perform any network calls (always true in Phase 1)")
	cmd.Flags().BoolVar(&share, "share", false, "upload sanitized analysis to sevro.dev/r/<hash> (opt-in)")
	cmd.Flags().BoolVar(&roast, "roast", false, "humorous output (findings stay accurate)")
	return cmd
}

// demoChart is the bundled demo values file. //go:embed lets us ship
// the fixture inside the binary so `npx @sevro/cli demo` works with
// no input.
//
//go:embed demo/values.yaml
var demoChart []byte

func newDemoCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Run analysis on a bundled demo chart",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rep, err := analyze.Run(bytesReader(demoChart), analyze.Options{Source: "demo"})
			if err != nil {
				return err
			}
			if jsonOut {
				return render.JSON(cmd.OutOrStdout(), rep)
			}
			return render.Text(cmd.OutOrStdout(), rep)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	return cmd
}

// bytesReader is a tiny adapter so analyze.Run can read from a byte slice
// without pulling in bytes.NewReader at the import-graph root of main.
func bytesReader(b []byte) *bytesReaderImpl { return &bytesReaderImpl{b: b} }

type bytesReaderImpl struct {
	b []byte
	i int
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <a> <b>",
		Short: "Show cost delta between two values files",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notYetImplemented(cmd)
		},
	}
}

func newScoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "score [chart]",
		Short: "Assign an efficiency score to a chart",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notYetImplemented(cmd)
		},
	}
}

func newAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "audit",
		Short:  "Audit a chart for security findings only",
		Hidden: false,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notYetImplemented(cmd)
		},
	}
}

func newWatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch [chart]",
		Short: "Watch a chart and re-analyze on change",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notYetImplemented(cmd)
		},
	}
}

func newCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare <a> <b>",
		Short: "Side-by-side comparison of two charts",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return notYetImplemented(cmd)
		},
	}
}

func notYetImplemented(cmd *cobra.Command) error {
	return fmt.Errorf("`sevro %s` is not yet implemented (see https://sevro.dev/roadmap)", cmd.Name())
}
