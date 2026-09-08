// Package gate implements phase-specific constitution compliance
// checking for the Unbound Force development workflow. It verifies
// that required spec artifacts, plans, tasks, coverage strategies,
// review markers, and branch discipline exist before allowing
// progression to the next workflow phase.
//
// Results are reported as structured data suitable for both
// colored terminal display and JSON output with provenance
// metadata per Constitution III.
package gate

// SchemaVersion is the JSON output schema version for backward
// compatibility tracking.
const SchemaVersion = "1.0.0"

// ProducerName identifies the tool that produced the gate report
// in provenance metadata per Constitution III.
const ProducerName = "uf-gate"

// CheckResult represents a single gate check finding.
// Gate checks are binary pass/fail — there is no warn severity
// because constitution compliance is not gradual.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

// Summary provides aggregate counts of check results.
type Summary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// GateReport is the complete output from a gate run, including
// provenance metadata per Constitution III.
type GateReport struct {
	Version   string        `json:"version"`
	Producer  string        `json:"producer"`
	Timestamp string        `json:"timestamp"`
	Branch    string        `json:"branch"`
	Phase     string        `json:"phase"`
	Passed    bool          `json:"passed"`
	Checks    []CheckResult `json:"checks"`
	Summary   Summary       `json:"summary"`
}
