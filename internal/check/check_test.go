package check

import (
	"strings"
	"testing"

	"github.com/roshbhatia/specutil/internal/ir"
)

// good builds a change that satisfies every rule in the rosh-spec-driven
// preset. Each test then breaks exactly one thing, so a finding can only come
// from the rule under test.
func good() *ir.Change {
	return &ir.Change{
		Name: "demo",
		Proposal: &ir.Proposal{
			Section: ir.Section{Raw: "## Why\n\nA reason.\n\n## What Changes\n\n- Do the thing\n\n### Non-goals\n\n- Not this\n"},
		},
		Design: &ir.Design{
			Section: ir.Section{Raw: "## Decisions\n\n- Decision: use X\n  - Alternative rejected: Y\n\n## Rollout & Gating\n\nBuild, then switch.\n\n## Adversarial Review\n\nPer the skill.\n"},
		},
		Specs: []*ir.Spec{{
			Section:    ir.Section{Raw: "## ADDED Requirements\n\n### Requirement: It works\n"},
			Capability: "cap",
			Requirements: []ir.Requirement{{
				Name:  "It works",
				Delta: ir.DeltaAdded,
				Scenarios: []ir.Scenario{
					{Name: "happy", Markers: map[string]string{"polarity": "positive"}},
					{Name: "sad", Markers: map[string]string{"polarity": "negative"}},
				},
			}},
		}},
		Tasks: &ir.Tasks{
			Section: ir.Section{Raw: "## 1. Build\n\n- [ ] 1.1 Do it\n- [ ] 1.2 Adversarial review\n"},
			Phases: []ir.Phase{
				{
					Number: "1", Name: "Build",
					Markers: map[string]string{"shape": "graph"},
					Items: []ir.TaskItem{
						{ID: "1.1", Text: "Do it"},
						{ID: "1.2", Text: "Adversarial review (skill)"},
					},
				},
				{
					Number: "2", Name: "Rollout",
					Items: []ir.TaskItem{{ID: "2.1", Text: "Apply: switch"}},
				},
			},
		},
	}
}

func roshRun(t *testing.T, c *ir.Change) *Report {
	t.Helper()
	rep, err := Run(Config{Preset: "rosh-spec-driven"}, []*ir.Change{c})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return rep
}

// rules returns the set of rule names that fired.
func firedRules(r *Report) map[string]bool {
	out := map[string]bool{}
	for _, f := range r.Findings {
		out[f.Rule] = true
	}
	return out
}

func TestPresetPassesAWellFormedChange(t *testing.T) {
	rep := roshRun(t, good())
	if !rep.OK() {
		t.Fatalf("expected a clean pass, got: %+v", rep.Findings)
	}
}

func TestMissingNonGoalsFails(t *testing.T) {
	c := good()
	c.Proposal.Raw = strings.Replace(c.Proposal.Raw, "### Non-goals\n\n- Not this\n", "", 1)
	if !firedRules(roshRun(t, c))["proposal-sections"] {
		t.Error("expected proposal-sections to fire")
	}
}

func TestMissingDesignSectionFails(t *testing.T) {
	c := good()
	c.Design.Raw = strings.Replace(c.Design.Raw, "## Adversarial Review", "## Something Else", 1)
	if !firedRules(roshRun(t, c))["design-sections"] {
		t.Error("expected design-sections to fire")
	}
}

func TestDecisionWithoutRejectedAlternativeFails(t *testing.T) {
	c := good()
	c.Design.Raw = strings.Replace(c.Design.Raw, "  - Alternative rejected: Y\n", "", 1)
	if !firedRules(roshRun(t, c))["paired-bullet"] {
		t.Error("expected paired-bullet to fire")
	}
}

// Two alternatives under one decision must not cover a bare decision elsewhere.
func TestPairedBulletCountsPerBlockNotInAggregate(t *testing.T) {
	c := good()
	c.Design.Raw = "## Decisions\n\n" +
		"- Decision: A\n  - Alternative rejected: a1\n  - Alternative rejected: a2\n" +
		"- Decision: B\n" +
		"\n## Rollout & Gating\n\nx\n\n## Adversarial Review\n\ny\n"
	if !firedRules(roshRun(t, c))["paired-bullet"] {
		t.Error("a bare second decision must fail even when the first has two alternatives")
	}
}

func TestRequirementWithoutNegativeScenarioFails(t *testing.T) {
	c := good()
	c.Specs[0].Requirements[0].Scenarios[1].Markers = map[string]string{"polarity": "positive"}
	if !firedRules(roshRun(t, c))["scenario-marker-coverage"] {
		t.Error("expected scenario-marker-coverage to fire")
	}
}

func TestRemovedRequirementNeedsNoNegativeScenario(t *testing.T) {
	c := good()
	c.Specs[0].Requirements = append(c.Specs[0].Requirements, ir.Requirement{
		Name: "Gone", Delta: ir.DeltaRemoved,
	})
	if !roshRun(t, c).OK() {
		t.Error("a removed requirement carries migration prose, not behavior, so it needs no scenario")
	}
}

func TestPhaseWithoutShapeFails(t *testing.T) {
	c := good()
	c.Tasks.Phases[0].Markers = nil
	if !firedRules(roshRun(t, c))["phase-marker-required"] {
		t.Error("expected phase-marker-required to fire")
	}
}

func TestRolloutPhaseIsExemptFromShape(t *testing.T) {
	// The Rollout phase in good() declares no shape and must still pass.
	if !roshRun(t, good()).OK() {
		t.Error("a rollout phase must not be required to declare a shape")
	}
}

func TestLoopPhaseWithoutStopAndCapFails(t *testing.T) {
	c := good()
	c.Tasks.Phases[0].Markers = map[string]string{"shape": "loop"}
	rep := roshRun(t, c)
	if !firedRules(rep)["phase-marker-conditional"] {
		t.Fatal("expected phase-marker-conditional to fire")
	}
	var msgs []string
	for _, f := range rep.Findings {
		msgs = append(msgs, f.Msg)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range []string{"stop", "maxIters"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected a finding naming %q, got:\n%s", want, joined)
		}
	}
}

func TestPhaseWithoutAdversarialReviewTaskFails(t *testing.T) {
	c := good()
	c.Tasks.Phases[0].Items = c.Tasks.Phases[0].Items[:1]
	if !firedRules(roshRun(t, c))["phase-task-pattern"] {
		t.Error("expected phase-task-pattern to fire")
	}
}

// A slice titled "Adversarial review" must not satisfy its own review gate; the
// rule reads task text, never the phase heading.
func TestPhaseTitleDoesNotSatisfyTaskPattern(t *testing.T) {
	c := good()
	c.Tasks.Phases[0].Name = "Adversarial review"
	c.Tasks.Phases[0].Items = []ir.TaskItem{{ID: "1.1", Text: "Do it"}}
	if !firedRules(roshRun(t, c))["phase-task-pattern"] {
		t.Error("a phase heading must not self-satisfy the review gate")
	}
}

func TestDanglingTaskDependencyFails(t *testing.T) {
	c := good()
	c.Tasks.Phases[0].Items[0].Fields = map[string][]string{"deps": {"9.9"}}
	if !firedRules(roshRun(t, c))["task-deps-resolve"] {
		t.Error("expected task-deps-resolve to fire")
	}
}

func TestEmDashFails(t *testing.T) {
	c := good()
	c.Proposal.Raw = strings.Replace(c.Proposal.Raw, "A reason.", "A reason — long.", 1)
	rep := roshRun(t, c)
	if !firedRules(rep)["no-em-dash"] {
		t.Fatal("expected no-em-dash to fire")
	}
	for _, f := range rep.Findings {
		if f.Rule == "no-em-dash" && f.Line == 0 {
			t.Error("an em-dash finding should carry the line it was found on")
		}
	}
}

func TestBoldedBulletLeadFails(t *testing.T) {
	c := good()
	c.Proposal.Raw = strings.Replace(c.Proposal.Raw, "- Do the thing", "- **Thing** does it", 1)
	if !firedRules(roshRun(t, c))["bolded-bullet-lead"] {
		t.Error("expected bolded-bullet-lead to fire")
	}
}

func TestAllowedBoldedBulletLeadsPass(t *testing.T) {
	c := good()
	c.Specs[0].Raw = "### Requirement: x\n\n#### Scenario: y\n" +
		"- **POLARITY** negative\n- **WHEN** a thing\n- **THEN** another\n"
	c.Tasks.Raw = "## 1. Build\n\n- **SHAPE** loop\n- **STOP** green\n- **MAX-ITERS** 3\n"
	if !roshRun(t, c).OK() {
		t.Errorf("the format keywords must be allowed as bolded leads: %+v", roshRun(t, c).Findings)
	}
}

func TestResolveRejectsUnknownPresetRuleAndSeverity(t *testing.T) {
	cases := map[string]Config{
		"unknown preset":   {Preset: "nope"},
		"unknown rule":     {Rules: []RuleConfig{{ID: "nope"}}},
		"missing id":       {Rules: []RuleConfig{{Severity: SeverityWarn}}},
		"unknown severity": {Rules: []RuleConfig{{ID: "no-em-dash", Severity: "loud"}}},
		"disable unknown":  {Preset: "rosh-spec-driven", Disable: []string{"nope"}},
	}
	for name, cfg := range cases {
		if _, err := Resolve(cfg); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestDisableRemovesARule(t *testing.T) {
	c := good()
	c.Proposal.Raw = strings.Replace(c.Proposal.Raw, "A reason.", "A reason — long.", 1)
	rep, err := Run(Config{Preset: "rosh-spec-driven", Disable: []string{"no-em-dash"}}, []*ir.Change{c})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.OK() {
		t.Errorf("disabling no-em-dash should clear the only violation, got %+v", rep.Findings)
	}
}

func TestSeverityOverrideDowngradesToWarning(t *testing.T) {
	c := good()
	c.Proposal.Raw = strings.Replace(c.Proposal.Raw, "A reason.", "A reason — long.", 1)
	rep, err := Run(Config{
		Preset: "rosh-spec-driven",
		Rules:  []RuleConfig{{ID: "no-em-dash", Severity: SeverityWarn}},
	}, []*ir.Change{c})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.OK() {
		t.Error("a warning must not fail the run")
	}
	if rep.Warnings() != 1 {
		t.Errorf("expected 1 warning, got %d", rep.Warnings())
	}
}

func TestLocalRuleOverridesPresetByName(t *testing.T) {
	c := good()
	c.Proposal.Raw = strings.Replace(c.Proposal.Raw, "- Do the thing", "- **Thing** does it", 1)
	rep, err := Run(Config{
		Preset: "rosh-spec-driven",
		Rules: []RuleConfig{{
			ID:     "bolded-bullet-lead",
			Params: map[string]any{"allow": []string{"Thing", "WHEN", "THEN", "POLARITY", "SHAPE", "STOP", "MAX-ITERS"}},
		}},
	}, []*ir.Change{c})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.OK() {
		t.Errorf("a local allow list should replace the preset's, got %+v", rep.Findings)
	}
}

func TestEmptyConfigChecksNothing(t *testing.T) {
	var zero Config
	if !zero.IsZero() {
		t.Error("the zero config must report itself as empty")
	}
	rep, err := Run(Config{}, []*ir.Change{good()})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Findings) != 0 {
		t.Errorf("no rubric means no findings, got %+v", rep.Findings)
	}
}

func TestReportOrderIsStable(t *testing.T) {
	c := good()
	c.Proposal.Raw = "- **A** x\n- **B** y\n"
	c.Design.Raw = ""
	first, err := Run(Config{Preset: "rosh-spec-driven"}, []*ir.Change{c})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		again, _ := Run(Config{Preset: "rosh-spec-driven"}, []*ir.Change{c})
		if len(again.Findings) != len(first.Findings) {
			t.Fatalf("finding count unstable: %d vs %d", len(again.Findings), len(first.Findings))
		}
		for j := range first.Findings {
			if again.Findings[j] != first.Findings[j] {
				t.Fatalf("finding %d unstable:\n%+v\n%+v", j, again.Findings[j], first.Findings[j])
			}
		}
	}
}

func TestAbsentOptionalArtifactIsNotAViolation(t *testing.T) {
	// A change with no design.md must not fail the design rules: whether the
	// artifact is required at all is the schema's call, not this rule's.
	c := good()
	c.Design = nil
	rep := roshRun(t, c)
	if fired := firedRules(rep); fired["design-sections"] || fired["paired-bullet"] {
		t.Errorf("absent design.md should not fire design rules, got %+v", rep.Findings)
	}
}
