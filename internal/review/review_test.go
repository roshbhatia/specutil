package review_test

import (
	"strings"
	"testing"

	"github.com/roshbhatia/specutil/internal/ident"
	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/review"
)

// change builds a minimal change whose one phase holds the given task texts.
func change(name string, tasks ...string) *ir.Change {
	items := make([]ir.TaskItem, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, ir.TaskItem{Text: t, Kind: ir.KindPlain})
	}
	return &ir.Change{
		Name:     name,
		Proposal: &ir.Proposal{Why: "because", Section: ir.Section{Raw: "## Why\n\nbecause\n"}},
		Tasks: &ir.Tasks{
			Section: ir.Section{Raw: "## 1. Build\n"},
			Phases:  []ir.Phase{{Number: "1", Name: "Build", Items: items}},
		},
	}
}

func idOf(t *testing.T, text string) string {
	t.Helper()
	return ident.Identity("Build", text)
}

func TestChangeHashIsStableAndSensitive(t *testing.T) {
	a := change("c", "do the thing")
	b := change("c", "do the thing")
	if review.ChangeHash(a) != review.ChangeHash(b) {
		t.Fatal("identical changes must hash identically")
	}
	b.Tasks.Raw = "## 1. Build\n\n- [ ] 1.1 do the thing\n"
	if review.ChangeHash(a) == review.ChangeHash(b) {
		t.Fatal("an edited artifact must change the hash")
	}
}

func TestUnreviewedChangeHasNoDrift(t *testing.T) {
	st := review.Build(change("c", "do the thing"), nil)
	if st.Reviewed {
		t.Fatal("a change with no record must not report as reviewed")
	}
	if !st.Gated() {
		t.Fatal("an unreviewed change must gate")
	}
	for _, is := range st.Items {
		if is.Drift != "" {
			t.Errorf("task %q reports drift %q with no baseline to compare against", is.Text, is.Drift)
		}
	}
}

func TestApplyThenBuildReportsCurrent(t *testing.T) {
	c := change("c", "do the thing")
	rec := review.Apply(c, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionApproved,
	})
	st := review.Build(c, rec)
	if st.Stale {
		t.Fatal("a record applied to the current change must not read as stale")
	}
	if st.Gated() {
		t.Fatal("a current approval must not gate")
	}
	for _, is := range st.Items {
		if is.Drift != review.DriftUnchanged {
			t.Errorf("task %q: got drift %q, want unchanged", is.Text, is.Drift)
		}
	}
}

func TestEditedArtifactMakesTheDecisionStale(t *testing.T) {
	c := change("c", "do the thing")
	rec := review.Apply(c, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionApproved,
	})
	c.Tasks.Raw += "- [ ] 1.2 and another\n"
	st := review.Build(c, rec)
	if !st.Stale {
		t.Fatal("editing an artifact after the decision must report stale")
	}
	if !st.Gated() {
		t.Fatal("a stale approval must gate")
	}
}

func TestAddedTaskIsNewAndRewordedTaskIsChanged(t *testing.T) {
	before := change("c", "wire up the token issuer")
	rec := review.Apply(before, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionApproved,
	})

	after := change("c", "wire up the token issuer with a configurable TTL", "add a metrics counter")
	drift := review.DriftByIdentity(after, rec)

	reworded := idOf(t, "wire up the token issuer with a configurable TTL")
	added := idOf(t, "add a metrics counter")

	if got := drift[reworded]; got != review.DriftChanged {
		t.Errorf("a reworded task: got %q, want %q", got, review.DriftChanged)
	}
	if got := drift[added]; got != review.DriftNew {
		t.Errorf("an added task: got %q, want %q", got, review.DriftNew)
	}
}

func TestCommentFollowsARewordedTask(t *testing.T) {
	before := change("c", "wire up the token issuer")
	rec := review.Apply(before, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionChangesRequested,
		Annotations: []review.Annotation{{
			Scope:    review.ScopeTask,
			Identity: idOf(t, "wire up the token issuer"),
			Comment:  "state the signing algorithm",
		}},
	})

	after := change("c", "wire up the token issuer with a configurable TTL")
	st := review.Build(after, rec)
	if len(st.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(st.Items))
	}
	if st.Items[0].Comment != "state the signing algorithm" {
		t.Errorf("the comment did not follow the reword: got %q", st.Items[0].Comment)
	}
}

func TestDropRequestSurvivesUntilTheTaskIsGone(t *testing.T) {
	c := change("c", "ship it on friday")
	rec := review.Apply(c, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionChangesRequested,
		Annotations: []review.Annotation{{
			Scope: review.ScopeTask, Identity: idOf(t, "ship it on friday"),
			Action: review.ActionDrop, Comment: "no friday deploys",
		}},
	})
	if got := len(review.Build(c, rec).Dropped); got != 1 {
		t.Fatalf("a requested removal on a present task must be reported: got %d", got)
	}
	if got := len(review.Build(change("c", "ship it on monday instead of friday"), rec).Dropped); got != 1 {
		t.Fatalf("a requested removal must follow a reword: got %d", got)
	}
	if got := len(review.Build(change("c", "run the migration"), rec).Dropped); got != 0 {
		t.Fatalf("a requested removal on a deleted task is resolved: got %d", got)
	}
}

func TestApplyDropsEmptyComments(t *testing.T) {
	c := change("c", "do the thing")
	rec := review.Apply(c, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionCommented,
		Annotations: []review.Annotation{
			{Scope: review.ScopeTask, Identity: idOf(t, "do the thing"), Comment: "   "},
			{Scope: review.ScopeTask, Identity: "abc", Comment: "keep me"},
		},
	})
	if len(rec.Annotations) != 1 || rec.Annotations[0].Comment != "keep me" {
		t.Fatalf("an empty comment must not be recorded: %+v", rec.Annotations)
	}
}

func TestValidateRejectsUnknownSchemaAndDecision(t *testing.T) {
	if err := (&review.Feedback{Schema: "nope", Decision: review.DecisionApproved}).Validate(); err == nil {
		t.Error("an unknown schema must be rejected rather than guessed at")
	}
	if err := (&review.Feedback{Schema: review.Schema, Decision: "lgtm"}).Validate(); err == nil {
		t.Error("an unknown decision must be rejected")
	}
	fb := &review.Feedback{
		Schema: review.Schema, Decision: review.DecisionApproved,
		Annotations: []review.Annotation{{Scope: review.ScopeTask, Comment: "x"}},
	}
	if err := fb.Validate(); err == nil {
		t.Error("a task annotation with no identity must be rejected")
	}
}

func TestMarkdownLeadsWithTheVerdict(t *testing.T) {
	c := change("c", "do the thing")
	rec := review.Apply(c, &review.Feedback{
		Schema: review.Schema, Change: "c", Decision: review.DecisionChangesRequested,
		Note: "split this up",
	})
	out := review.Markdown(review.Build(c, rec))
	if !strings.Contains(out, "Decision: changes-requested") {
		t.Errorf("the brief must state the decision: %s", out)
	}
	if !strings.Contains(out, "split this up") {
		t.Errorf("the brief must carry the note: %s", out)
	}
}
