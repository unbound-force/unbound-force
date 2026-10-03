package metrics

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/unbound-force/unbound-force/internal/artifacts"
)

const (
	reviewActionNone                   = "none"
	reviewActionAddressFindings        = "address-findings"
	reviewActionHumanResolution        = "human-resolution"
	reviewActionRerunOrHumanResolution = "successful-rerun-or-human-resolution"
)

type reviewVerdictMetric struct {
	SchemaVersion               string                          `json:"schema_version"`
	CouncilDecision             artifacts.ReviewVerdictDecision `json:"council_decision"`
	AutomatedProgressionAllowed bool                            `json:"automated_progression_allowed"`
	RequiredAction              string                          `json:"required_action"`
}

// CollectDivisor collects metrics from Divisor review verdict artifacts.
func CollectDivisor(artifactDir string, since time.Time) (*SourceCollection, error) {
	paths, err := artifacts.FindArtifacts(artifactDir, "review-verdict")
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	raw := make(map[string]interface{})
	var verdicts []map[string]interface{}
	var decisions []reviewVerdictMetric
	skippedVerdicts := 0

	for _, p := range paths {
		env, err := artifacts.ReadEnvelope(p)
		if err != nil {
			skippedVerdicts++
			continue
		}
		payload, decision, err := readReviewVerdict(env)
		if err != nil {
			skippedVerdicts++
			continue
		}
		verdicts = append(verdicts, payload)
		decisions = append(decisions, decision)
	}

	raw["verdicts"] = verdicts
	raw["verdict_count"] = len(verdicts)
	raw["review_decisions"] = decisions
	raw["skipped_verdict_count"] = skippedVerdicts

	return &SourceCollection{
		Source:      "divisor",
		CollectedAt: now,
		DataPoints:  len(verdicts),
		RawData:     raw,
	}, nil
}

func readReviewVerdict(env *artifacts.Envelope) (map[string]interface{}, reviewVerdictMetric, error) {
	consumerVersion, err := supportedReviewVerdictVersion(env.SchemaVersion)
	if err != nil {
		return nil, reviewVerdictMetric{}, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		return nil, reviewVerdictMetric{}, fmt.Errorf("decode review-verdict payload: %w", err)
	}

	var decision artifacts.ReviewVerdictDecision
	switch consumerVersion {
	case artifacts.ReviewVerdictSchemaVersionV1, artifacts.ReviewVerdictSchemaVersionV2:
		decisionValue, ok := payload["council_decision"].(string)
		if !ok {
			return nil, reviewVerdictMetric{}, fmt.Errorf(
				"decode review-verdict decision for schema %q: expected string",
				env.SchemaVersion,
			)
		}
		decision = artifacts.ReviewVerdictDecision(decisionValue)
	default:
		return nil, reviewVerdictMetric{}, fmt.Errorf("unsupported review-verdict consumer version %q", consumerVersion)
	}

	metric, err := classifyReviewVerdict(consumerVersion, env.SchemaVersion, decision)
	if err != nil {
		return nil, reviewVerdictMetric{}, err
	}
	return payload, metric, nil
}

func supportedReviewVerdictVersion(producerVersion string) (string, error) {
	envelope := &artifacts.Envelope{SchemaVersion: producerVersion}
	v2Compatible, v2Warning := artifacts.CheckSchemaVersion(
		envelope,
		artifacts.ReviewVerdictSchemaVersionV2,
	)
	if v2Compatible {
		return artifacts.ReviewVerdictSchemaVersionV2, nil
	}

	v1Compatible, v1Warning := artifacts.CheckSchemaVersion(
		envelope,
		artifacts.ReviewVerdictSchemaVersionV1,
	)
	if v1Compatible {
		return artifacts.ReviewVerdictSchemaVersionV1, nil
	}

	return "", fmt.Errorf(
		"unsupported review-verdict schema version %q: v2 compatibility: %v; v1 compatibility: %v",
		producerVersion,
		v2Warning,
		v1Warning,
	)
}

func classifyReviewVerdict(
	consumerVersion string,
	producerVersion string,
	decision artifacts.ReviewVerdictDecision,
) (reviewVerdictMetric, error) {
	metric := reviewVerdictMetric{
		SchemaVersion:   producerVersion,
		CouncilDecision: decision,
	}

	switch decision {
	case artifacts.ReviewVerdictApproved:
		metric.AutomatedProgressionAllowed = true
		metric.RequiredAction = reviewActionNone
	case artifacts.ReviewVerdictChangesRequested:
		metric.RequiredAction = reviewActionAddressFindings
	case artifacts.ReviewVerdictEscalated:
		metric.RequiredAction = reviewActionHumanResolution
	case artifacts.ReviewVerdictInconclusive, artifacts.ReviewVerdictUnavailable:
		if consumerVersion == artifacts.ReviewVerdictSchemaVersionV1 {
			return reviewVerdictMetric{}, fmt.Errorf(
				"review-verdict decision %q is invalid for historical schema %q",
				decision,
				producerVersion,
			)
		}
		metric.RequiredAction = reviewActionRerunOrHumanResolution
	default:
		return reviewVerdictMetric{}, fmt.Errorf(
			"unknown review-verdict decision %q for schema %q",
			decision,
			producerVersion,
		)
	}

	return metric, nil
}
