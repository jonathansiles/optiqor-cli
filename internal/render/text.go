// Package render formats analysis results as styled text or JSON.
//
// **Every renderer must include the ±40% accuracy disclosure.** Removing
// it is a hard rule violation — see ../../CLAUDE.md. The disclosure is
// what makes the CLI a trustworthy funnel: we never overpromise.
//
// Layout philosophy: cost is the headline product, security findings
// are a bonus side-effect of parsing. Renderers therefore split
// findings by [rules.Category] and present them as two distinct
// sections — cost first with full detail, security after as compact
// one-liners — so a user scanning the output sees the value prop in
// the first screenful.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/optiqor/optiqor-cli/internal/render/style"
	"github.com/optiqor/optiqor-cli/pkg/rules"
)

// AccuracyDisclosure is the mandatory line every output must contain.
const AccuracyDisclosure = "Sandbox accuracy: ±40%. Install the Optiqor agent for exact numbers (optiqor.dev/get)."

// Brand strings used in the header banner.
const (
	BrandName    = "optiqor"
	BrandTagline = "Helm chart cost optimization · security as a bonus"
	GetURL       = "https://optiqor.dev/get"
)

// Layout constants. Centralised so callers don't sprinkle magic
// numbers and so visual tweaks happen in one place.
const (
	defaultWidth     = 78
	contentIndent    = "  "
	findingIndent    = "    "
	monthsPerYear    = 12
	annualTeaserMin  = 1_00 // show annual projection only above $1/mo savings
	bonusSectionName = "Security findings"
	costSectionName  = "Cost optimizations"
)

// Report is the renderer-facing view of an analysis run.
type Report struct {
	Source    string          `json:"source"` // path or label of the input
	Workloads int             `json:"workloads_analyzed"`
	Findings  []rules.Finding `json:"findings"`
}

// Options controls how a Report is rendered. Callers (cmd/optiqor/main.go)
// detect TTY + NO_COLOR + --no-color and set Color accordingly.
type Options struct {
	Color bool // false → plain ASCII, no ANSI; true → branded styled output
	Width int  // terminal width; 0 → defaultWidth
}

// MonthlySavingsUSDCents totals the predicted savings across findings.
func (r Report) MonthlySavingsUSDCents() int64 {
	var sum int64
	for _, f := range r.Findings {
		sum += f.MonthlyUSDCents
	}
	return sum
}

// Text writes the styled human-readable report. The output is split
// into a branded header, an executive summary, a "Cost optimizations"
// section (full detail), a "Security findings (bonus)" section
// (compact one-liners), and a footer with the accuracy disclosure
// and agent CTA.
func Text(w io.Writer, r Report, opts Options) error {
	t := style.NewTheme(opts.Color)
	width := opts.Width
	if width <= 0 {
		width = defaultWidth
	}

	cost, security := splitByCategory(r.Findings)

	var b strings.Builder
	writeHeader(&b, t, width)
	writeSummary(&b, t, r, len(cost), len(security))

	if len(cost) == 0 && len(security) == 0 {
		fmt.Fprintf(&b, "\n%s%s\n\n", contentIndent, t.OK.Render("✓ Clean. No findings."))
		writeFooter(&b, t, width, 0)
		_, err := io.WriteString(w, b.String())
		return err
	}

	if len(cost) > 0 {
		writeCostSection(&b, t, width, sortCostForDisplay(cost))
	}
	if len(security) > 0 {
		writeSecuritySection(&b, t, width, security)
	}

	writeFooter(&b, t, width, r.MonthlySavingsUSDCents())
	_, err := io.WriteString(w, b.String())
	return err
}

// sortCostForDisplay reorders cost findings so the biggest dollar
// impact leads. The engine's stable sort (workload → severity) is
// excellent for diffs and audit, but it buries high-savings findings
// behind alphabetically-earlier workloads. Display order:
//
//  1. findings with monthly savings, highest USD first
//  2. then findings with no dollar estimate, by severity desc
//  3. ties broken by workload, then detector ID — both stable.
//
// Returns a new slice; the caller's input is not mutated.
func sortCostForDisplay(in []rules.Finding) []rules.Finding {
	out := make([]rules.Finding, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.MonthlyUSDCents > 0) != (b.MonthlyUSDCents > 0) {
			return a.MonthlyUSDCents > b.MonthlyUSDCents
		}
		if a.MonthlyUSDCents != b.MonthlyUSDCents {
			return a.MonthlyUSDCents > b.MonthlyUSDCents
		}
		if a.Severity != b.Severity {
			return severityRank(a.Severity) > severityRank(b.Severity)
		}
		if a.Workload != b.Workload {
			return a.Workload < b.Workload
		}
		return a.DetectorID < b.DetectorID
	})
	return out
}

func severityRank(s rules.Severity) int {
	switch s {
	case rules.SeverityHigh:
		return 3
	case rules.SeverityMed:
		return 2
	case rules.SeverityLow:
		return 1
	}
	return 0
}

// splitByCategory partitions findings while preserving the order
// established by rules.Run (workload → severity → detector ID).
// Findings without a Category fall back to the cost section so they
// remain visible — a defensive default for any custom detector that
// forgets to declare one.
func splitByCategory(findings []rules.Finding) (cost, security []rules.Finding) {
	cost = make([]rules.Finding, 0, len(findings))
	security = make([]rules.Finding, 0, len(findings))
	for _, f := range findings {
		switch f.Category {
		case rules.CategorySecurity:
			security = append(security, f)
		default:
			cost = append(cost, f)
		}
	}
	return cost, security
}

func writeHeader(b *strings.Builder, t style.Theme, width int) {
	div := t.DividerLine(width)
	mark := t.BrandMark.Render(style.BrandGlyph)
	brand := t.Brand.Render(BrandName)
	tag := t.Tagline.Render(BrandTagline)
	fmt.Fprintf(b, "%s\n", div)
	fmt.Fprintf(b, "%s%s  %s\n", contentIndent, mark, brand)
	fmt.Fprintf(b, "%s%s\n", contentIndent, tag)
	fmt.Fprintf(b, "%s\n\n", div)
}

func writeSummary(b *strings.Builder, t style.Theme, r Report, costCount, secCount int) {
	srcLabel := r.Source
	if srcLabel == "" {
		srcLabel = "(stdin)"
	}
	monthly := r.MonthlySavingsUSDCents()

	rows := [][2]string{
		{"Source", t.Workload.Render(srcLabel)},
		{"Workloads", t.Title.Render(plural(r.Workloads, "workload", "workloads")) + t.Muted.Render(" analyzed")},
		{"Cost", costSummaryValue(t, costCount, monthly)},
	}
	if secCount > 0 {
		rows = append(rows, [2]string{
			"Security",
			t.Muted.Render(plural(secCount, "finding", "findings")+" — bonus, surfaced while parsing"),
		})
	}

	labelWidth := 0
	for _, row := range rows {
		if n := len([]rune(row[0])); n > labelWidth {
			labelWidth = n
		}
	}
	for _, row := range rows {
		label := row[0] + strings.Repeat(" ", labelWidth-len([]rune(row[0])))
		fmt.Fprintf(b, "%s%s   %s\n", contentIndent, t.Muted.Render(label), row[1])
	}
}

func costSummaryValue(t style.Theme, count int, monthlyCents int64) string {
	if count == 0 {
		return t.OK.Render("✓ no cost waste detected")
	}
	primary := t.Title.Render(plural(count, "optimization", "optimizations"))
	if monthlyCents <= 0 {
		return primary
	}
	monthly := t.BigSavings.Render("save ~$" + formatCents(monthlyCents) + "/mo")
	suffix := ""
	if monthlyCents >= annualTeaserMin {
		suffix = t.Muted.Render(fmt.Sprintf(" (~$%s/yr)", formatCents(monthlyCents*monthsPerYear)))
	}
	return primary + t.Muted.Render(" · ") + monthly + suffix + t.Muted.Render(" ±40%")
}

func writeCostSection(b *strings.Builder, t style.Theme, width int, findings []rules.Finding) {
	b.WriteString("\n")
	fmt.Fprintf(b, "%s\n\n", t.SectionRule(costSectionName, width, t.SectionPrimary))
	for i, f := range findings {
		writeCostFinding(b, t, f, width)
		if i < len(findings)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
}

func writeCostFinding(b *strings.Builder, t style.Theme, f rules.Finding, width int) {
	badge := t.SeverityBadge(string(f.Severity))
	wl := t.Workload.Render(f.Workload)

	if f.MonthlyUSDCents > 0 {
		savings := t.Savings.Render("save ~$" + formatCents(f.MonthlyUSDCents) + "/mo")
		fmt.Fprintf(b, "%s%s  %s   %s\n", contentIndent, badge, wl, savings)
	} else {
		fmt.Fprintf(b, "%s%s  %s\n", contentIndent, badge, wl)
	}

	title := f.Title
	if title == "" && f.DetectorID != "" {
		title = f.DetectorID
	}
	fmt.Fprintf(b, "%s%s\n", findingIndent, t.Title.Render(title))
	for _, line := range wrap(f.Detail, width-len(findingIndent)) {
		fmt.Fprintf(b, "%s%s\n", findingIndent, t.Detail.Render(line))
	}
	fmt.Fprintf(b, "%s%s %s\n", findingIndent,
		t.Muted.Render("confidence:"),
		t.ConfidenceDots(string(f.Confidence)),
	)
}

func writeSecuritySection(b *strings.Builder, t style.Theme, width int, findings []rules.Finding) {
	label := fmt.Sprintf("%s  (bonus, %d)", bonusSectionName, len(findings))
	b.WriteString("\n")
	fmt.Fprintf(b, "%s\n", t.SectionRule(label, width, t.SectionBonus))
	fmt.Fprintf(b, "%s%s\n\n", contentIndent,
		t.SectionSubtle.Render("Spotted while parsing your chart. Cost is the headline; this is a bonus."),
	)

	maxWorkload := 0
	for _, f := range findings {
		if n := len([]rune(f.Workload)); n > maxWorkload {
			maxWorkload = n
		}
	}
	if maxWorkload > 24 { // keep the column reasonable on narrow terminals
		maxWorkload = 24
	}

	for _, f := range findings {
		writeSecurityFinding(b, t, f, maxWorkload)
	}
	b.WriteString("\n")
	fmt.Fprintf(b, "%s%s\n\n", contentIndent,
		t.Muted.Render("Run `optiqor audit` to focus only on these findings."),
	)
}

func writeSecurityFinding(b *strings.Builder, t style.Theme, f rules.Finding, workloadColWidth int) {
	wl := truncate(f.Workload, workloadColWidth)
	wlPadded := wl + strings.Repeat(" ", workloadColWidth-len([]rune(wl)))
	title := f.Title
	if title == "" && f.DetectorID != "" {
		title = f.DetectorID
	}
	fmt.Fprintf(b, "%s%s  %s   %s   %s\n",
		contentIndent,
		t.SeverityBadge(string(f.Severity)),
		t.Workload.Render(wlPadded),
		t.ConfidenceGlyph(string(f.Confidence)),
		t.Detail.Render(title),
	)
}

func writeFooter(b *strings.Builder, t style.Theme, width int, totalCents int64) {
	fmt.Fprintf(b, "%s\n", t.DividerLine(width))
	if totalCents > 0 {
		fmt.Fprintf(b, "%s%s %s   %s\n", contentIndent,
			t.Muted.Render("estimated monthly savings:"),
			t.BigSavings.Render("$"+formatCents(totalCents)+"/mo"),
			t.Muted.Render("(±40%)"),
		)
	}
	fmt.Fprintf(b, "%s%s\n", contentIndent, t.Disclosure.Render(AccuracyDisclosure))
	linkLabel := t.CallToLink.Render("optiqor.dev/get")
	fmt.Fprintf(b, "%s%s %s\n", contentIndent,
		t.Muted.Render("→ install the agent for exact numbers:"),
		t.Hyperlink(linkLabel, GetURL),
	)
}

// JSON writes the report as machine-readable JSON. Always disclosure-
// gated. Never colored — JSON output is for piping. The schema groups
// findings by category so consumers don't have to replicate the
// renderer's split.
func JSON(w io.Writer, r Report) error {
	cost, security := splitByCategory(r.Findings)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(struct {
		AccuracyDisclosure string          `json:"accuracy_disclosure"`
		Source             string          `json:"source"`
		Workloads          int             `json:"workloads_analyzed"`
		Findings           []rules.Finding `json:"findings"`
		Cost               []rules.Finding `json:"cost_findings"`
		Security           []rules.Finding `json:"security_findings_bonus"`
		MonthlySavingsUSD  float64         `json:"monthly_savings_usd"`
		AnnualSavingsUSD   float64         `json:"annual_savings_usd"`
	}{
		AccuracyDisclosure: AccuracyDisclosure,
		Source:             r.Source,
		Workloads:          r.Workloads,
		Findings:           r.Findings,
		Cost:               cost,
		Security:           security,
		MonthlySavingsUSD:  float64(r.MonthlySavingsUSDCents()) / 100.0,
		AnnualSavingsUSD:   float64(r.MonthlySavingsUSDCents()*monthsPerYear) / 100.0,
	})
}

func formatCents(c int64) string {
	dollars := c / 100
	cents := c % 100
	if cents == 0 {
		return fmt.Sprintf("%d", dollars)
	}
	return fmt.Sprintf("%d.%02d", dollars, cents)
}

func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralForm)
}

// wrap breaks s into lines no wider than width runes. Naïve word wrap;
// good enough for finding details which are sentences, not paragraphs.
func wrap(s string, width int) []string {
	if width <= 0 || len([]rune(s)) <= width {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	words := strings.Fields(s)
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len() == 0 {
			cur.WriteString(w)
			continue
		}
		if cur.Len()+1+len(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
			continue
		}
		cur.WriteString(" ")
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

// truncate clamps s to at most width runes, suffixing "…" when it had
// to cut. Width must be >= 1; callers pass column widths so this is
// trivially satisfied in practice.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(r[:width-1]) + "…"
}
