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
			ID: "phase-marker-required",
			Params: map[string]any{
				"marker":        "shape",
				"allowedValues": []string{"loop", "graph"},
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
		{
			// The schema's own words: "STOP is a command, not a wish." A prose STOP
			// cannot be evaluated, so the loop ends when the model decides it is
			// done, and `loop-gate arm --until` has nothing to run.
			ID:   "phase-marker-pattern",
			Name: "stop-is-a-command",
			Params: map[string]any{
				"marker": "stop",
				"when":   map[string]any{"marker": "shape", "value": "loop"},
				// Any backtick span is too weak: a stop that says "`lib/x.nix` passes
				// fixtures" names a file, not something to run, and passed. Require the
				// span to open with a command. The vocabulary lives here, in the
				// framework's preset, so the rule itself stays framework-agnostic.
				"pattern":  "`(nix|nh|task|specutil|openspec|citelock|loop-gate|git|jq|go|python3?|bash|sh|shellcheck|shfmt|pytest|npm|bun|cargo|make)\\b[^`]*`",
				"describe": "open a backtick span with a command",
			},
		},
		{ID: "task-deps-resolve"},
		{ID: "task-id-required"},
		{ID: "task-deps-acyclic"},
		{ID: "no-em-dash"},
		{
			ID: "bolded-bullet-lead",
			Params: map[string]any{
				"allow": []string{
					// POLARITY stays allowed: no preset requires a polarity marker
					// any more, but archived and other-repo spec files still carry
					// one, and rewriting them would be churn for no signal.
					"WHEN", "THEN", "AND", "POLARITY", "SHAPE", "STOP", "MAX-ITERS", "BREAKING",
				},
			},
		},
		{
			// A change nobody approved is not ready, and an approval given before
			// the artifacts moved is not an approval of what is there now. Both
			// read as a pass to every other rule in this preset, so the gate has
			// to state it. A repository that reviews out of band drops this with
			// `disable: [review-decision-current]`.
			ID: "review-decision-current",
			Params: map[string]any{
				"accept": []string{"approved"},
			},
		},
	},
}
