package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReviewVerdictConsumers_AgentMigrationContracts(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	cobaltPath := filepath.Join(root, ".opencode", "agents", "cobalt-crush-dev.md")
	cobalt, err := os.ReadFile(cobaltPath)
	if err != nil {
		t.Fatalf("read Cobalt-Crush consumer %s: %v", cobaltPath, err)
	}
	embeddedCobalt, err := assetContent("opencode/agents/cobalt-crush-dev.md")
	if err != nil {
		t.Fatalf("read scaffolded Cobalt-Crush consumer: %v", err)
	}

	mutiMindPath := filepath.Join(root, ".opencode", "agents", "muti-mind-po.md")
	mutiMind, err := os.ReadFile(mutiMindPath)
	if err != nil {
		t.Fatalf("read Muti-Mind consumer %s: %v", mutiMindPath, err)
	}

	consumers := []commandCopy{
		{name: "Cobalt-Crush canonical", text: string(cobalt)},
		{name: "Cobalt-Crush scaffold", text: string(embeddedCobalt)},
		{name: "Muti-Mind", text: string(mutiMind)},
	}
	commonClauses := []commandClause{
		{
			name: "canonical artifact and versions",
			all: []string{
				"`review-verdict` as the canonical decision artifact",
				"historical schema version 1 and current schema version 2",
				"same-major compatibility rule",
				"an unmigrated v1-only consumer must reject version 2",
			},
		},
		{
			name: "approval",
			all:  []string{"`approved` permits", "progression"},
		},
		{
			name: "blocking",
			all:  []string{"`changes_requested` blocks automated progression", "findings are addressed"},
		},
		{
			name: "advisory",
			all:  []string{"`escalated` is advisory", "blocks automated progression", "human gate"},
		},
		{
			name: "no-success",
			all: []string{
				"`inconclusive` and `unavailable` are not approvals",
				"both block automated progression",
				"successful rerun or explicit human resolution",
			},
		},
	}

	for _, consumer := range consumers {
		consumer := consumer
		t.Run(consumer.name, func(t *testing.T) {
			normalized := normalizeCommandText(consumer.text)
			for _, clause := range commonClauses {
				for _, fragment := range clause.all {
					assertNormalizedContains(t, consumer.name, normalized, clause.name, fragment)
				}
			}
		})
	}
}
