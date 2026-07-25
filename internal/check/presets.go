package check

// presets are built-in rubrics, keyed by the spec-framework schema name they
// describe. A preset is data: a bundle of built-in rules with the parameters
// that framework expects. Nothing in the rule implementations knows a framework
// name, so a new framework is a new entry here and nothing else.
var presets = map[string][]RuleConfig{
	// rosh-spec-driven mirrors the rubric the specreview shell lint enforced,
	// rule for rule, so a repository can drop that script and get the same
	// verdicts from specutil.
	"rosh-spec-driven": {
		{
			ID:   "required-sections",
			Name: "proposal-sections",
			Params: map[string]any{
				"artifact": "proposal",
				"sections": []string{"### Non-goals"},
			},
		},
		{
			ID:   "required-sections",
			Name: "design-sections",
			Params: map[string]any{
				"artifact": "design",
				"sections": []string{"## Decisions", "## Rollout & Gating", "## Adversarial Review"},
			},
		},
		{
			ID: "paired-bullet",
			Params: map[string]any{
				"artifact": "design",
				"lead":     "- Decision:",
				"follower": "- Alternative rejected:",
			},
		},
		{
			ID: "scenario-marker-coverage",
			Params: map[string]any{
				"marker": "polarity",
				"value":  "negative",
			},
		},
		{
			ID: "phase-marker-required",
			Params: map[string]any{
				"marker": "shape",
				// A rollout slice sequences the impactful actions; it is exempt
				// from declaring a work shape.
				"skipPhasePattern": "(?i)rollout",
			},
		},
		{
			ID: "phase-marker-conditional",
			Params: map[string]any{
				"when":             map[string]any{"marker": "shape", "value": "loop"},
				"require":          []string{"stop", "maxIters"},
				"skipPhasePattern": "(?i)rollout",
			},
		},
		{
			ID: "phase-task-pattern",
			Params: map[string]any{
				"pattern":          `(?i)adversarial\s+review`,
				"describe":         "adversarial-review task",
				"skipPhasePattern": "(?i)rollout",
			},
		},
		{ID: "task-deps-resolve"},
		{ID: "no-em-dash"},
		{
			ID: "bolded-bullet-lead",
			Params: map[string]any{
				"allow": []string{
					"WHEN", "THEN", "AND", "POLARITY", "SHAPE", "STOP", "MAX-ITERS", "BREAKING",
				},
			},
		},
	},
}
