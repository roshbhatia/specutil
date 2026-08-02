package check

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/roshbhatia/specutil/internal/ir"
	"github.com/roshbhatia/specutil/internal/review"
)

func init() {
	register(rule{
		id:  "required-sections",
		doc: "an artifact must contain each named heading (params: artifact, sections)",
		eval: func(p params, c *ir.Change) []Finding {
			artifact := p.String("artifact")
			text, file, ok := artifactText(c, artifact)
			if !ok {
				return nil
			}
			var out []Finding
			for _, want := range p.Strings("sections") {
				if line := findLine(text, func(l string) bool { return strings.TrimSpace(l) == want }); line == 0 {
					out = append(out, Finding{
						File: file,
						Msg:  fmt.Sprintf("%s is missing the required section %q", file, want),
					})
				}
			}
			return out
		},
	})

	register(rule{
		id: "paired-bullet",
		doc: "every bullet matching lead must be followed by one matching follower " +
			"before the next lead (params: artifact, lead, follower)",
		eval: func(p params, c *ir.Change) []Finding {
			text, file, ok := artifactText(c, p.String("artifact"))
			if !ok {
				return nil
			}
			lead, follower := p.String("lead"), p.String("follower")
			if lead == "" || follower == "" {
				return nil
			}
			var out []Finding
			openLine, openText := 0, ""
			paired := true
			flush := func() {
				if openLine != 0 && !paired {
					out = append(out, Finding{
						File: file, Line: openLine,
						Msg: fmt.Sprintf("%q has no following %q", openText, follower),
					})
				}
			}
			for i, raw := range strings.Split(text, "\n") {
				line := strings.TrimSpace(raw)
				switch {
				case strings.HasPrefix(line, lead):
					flush()
					openLine, openText, paired = i+1, line, false
				case strings.HasPrefix(line, follower):
					if openLine != 0 {
						paired = true
					}
				}
			}
			flush()
			return out
		},
	})

	register(rule{
		id: "scenario-marker-coverage",
		doc: "every requirement must have at least one scenario declaring marker=value " +
			"(params: marker, value)",
		eval: func(p params, c *ir.Change) []Finding {
			marker, value := p.String("marker"), p.String("value")
			if marker == "" {
				return nil
			}
			var out []Finding
			for _, s := range c.Specs {
				if s == nil {
					continue
				}
				file := "specs/" + s.Capability + "/spec.md"
				for _, r := range s.Requirements {
					// A removed or renamed requirement carries migration prose,
					// not behavior, so it has no scenarios to cover.
					if r.Delta == ir.DeltaRemoved || r.Delta == ir.DeltaRenamed {
						continue
					}
					found := false
					for _, sc := range r.Scenarios {
						if got, ok := sc.Markers[marker]; ok && strings.EqualFold(got, value) {
							found = true
							break
						}
					}
					if !found {
						out = append(out, Finding{
							File: file,
							Msg: fmt.Sprintf("requirement %q has no scenario declaring %s=%s",
								r.Name, marker, value),
						})
					}
				}
			}
			return out
		},
	})

	register(rule{
		id:  "phase-marker-required",
		doc: "every phase must declare the named marker (params: marker, skipPhasePattern)",
		eval: func(p params, c *ir.Change) []Finding {
			marker := p.String("marker")
			if marker == "" || c.Tasks == nil {
				return nil
			}
			skip := compile(p.String("skipPhasePattern"))
			var out []Finding
			for _, ph := range c.Tasks.Phases {
				if skipped(skip, ph) {
					continue
				}
				if _, ok := ph.Markers[marker]; !ok {
					out = append(out, Finding{
						File: "tasks.md",
						Msg:  fmt.Sprintf("phase %q declares no %s marker", phaseLabel(ph), marker),
					})
				}
			}
			return out
		},
	})

	register(rule{
		id: "phase-marker-conditional",
		doc: "a phase declaring when.marker=when.value must also declare each required marker " +
			"(params: when {marker, value}, require, skipPhasePattern)",
		eval: func(p params, c *ir.Change) []Finding {
			when, _ := p["when"].(map[string]any)
			if when == nil || c.Tasks == nil {
				return nil
			}
			trigger := params(when).String("marker")
			value := params(when).String("value")
			required := p.Strings("require")
			if trigger == "" || len(required) == 0 {
				return nil
			}
			skip := compile(p.String("skipPhasePattern"))
			var out []Finding
			for _, ph := range c.Tasks.Phases {
				if skipped(skip, ph) {
					continue
				}
				got, ok := ph.Markers[trigger]
				if !ok || !strings.EqualFold(got, value) {
					continue
				}
				for _, need := range required {
					if _, ok := ph.Markers[need]; !ok {
						out = append(out, Finding{
							File: "tasks.md",
							Msg: fmt.Sprintf("phase %q is %s=%s but declares no %s marker",
								phaseLabel(ph), trigger, value, need),
						})
					}
				}
			}
			return out
		},
	})

	register(rule{
		id: "phase-task-pattern",
		doc: "every phase must contain a task matching the pattern " +
			"(params: pattern, skipPhasePattern, describe)",
		eval: func(p params, c *ir.Change) []Finding {
			re := compile(p.String("pattern"))
			if re == nil || c.Tasks == nil {
				return nil
			}
			describe := p.String("describe")
			if describe == "" {
				describe = "a task matching " + p.String("pattern")
			}
			skip := compile(p.String("skipPhasePattern"))
			var out []Finding
			for _, ph := range c.Tasks.Phases {
				if skipped(skip, ph) {
					continue
				}
				found := false
				for _, it := range ph.Items {
					if re.MatchString(it.Text) {
						found = true
						break
					}
				}
				if !found {
					out = append(out, Finding{
						File: "tasks.md",
						Msg:  fmt.Sprintf("phase %q has no %s", phaseLabel(ph), describe),
					})
				}
			}
			return out
		},
	})

	register(rule{
		id:  "task-deps-resolve",
		doc: "every declared task dependency must name a task in the same change",
		eval: func(_ params, c *ir.Change) []Finding {
			if c.Tasks == nil {
				return nil
			}
			known := map[string]bool{}
			for _, ph := range c.Tasks.Phases {
				for _, it := range ph.Items {
					if it.ID != "" {
						known[it.ID] = true
					}
				}
			}
			var out []Finding
			for _, ph := range c.Tasks.Phases {
				for _, it := range ph.Items {
					for _, dep := range it.Fields["deps"] {
						if !known[dep] {
							out = append(out, Finding{
								File: "tasks.md",
								Msg:  fmt.Sprintf("task %s depends on %q, which names no task in this change", it.ID, dep),
							})
						}
					}
				}
			}
			return out
		},
	})

	register(rule{
		id:  "task-deps-acyclic",
		doc: "declared task dependencies must not form a cycle",
		eval: func(_ params, c *ir.Change) []Finding {
			if c.Tasks == nil {
				return nil
			}
			deps := map[string][]string{}
			var order []string
			for _, ph := range c.Tasks.Phases {
				for _, it := range ph.Items {
					if it.ID == "" {
						continue
					}
					if _, seen := deps[it.ID]; !seen {
						order = append(order, it.ID)
					}
					deps[it.ID] = append(deps[it.ID], it.Fields["deps"]...)
				}
			}
			const (
				unvisited = iota
				onStack
				done
			)
			state := map[string]int{}
			var stack []string
			var out []Finding
			var walk func(string)
			walk = func(id string) {
				state[id] = onStack
				stack = append(stack, id)
				for _, dep := range deps[id] {
					if _, known := deps[dep]; !known {
						continue // task-deps-resolve already reports the dangling edge
					}
					switch state[dep] {
					case unvisited:
						walk(dep)
					case onStack:
						out = append(out, Finding{
							File: "tasks.md",
							Msg: fmt.Sprintf("task dependencies form a cycle: %s",
								strings.Join(append(cycleFrom(stack, dep), dep), " -> ")),
						})
					}
				}
				stack = stack[:len(stack)-1]
				state[id] = done
			}
			for _, id := range order {
				if state[id] == unvisited {
					walk(id)
				}
			}
			return out
		},
	})

	register(rule{
		id:  "no-em-dash",
		doc: "no artifact may contain an em-dash",
		eval: func(_ params, c *ir.Change) []Finding {
			var out []Finding
			for _, a := range allArtifacts(c) {
				if line := findLine(a.Text, func(l string) bool { return strings.ContainsRune(l, '—') }); line != 0 {
					out = append(out, Finding{
						File: a.File, Line: line,
						Msg: fmt.Sprintf("%s contains an em-dash; use a comma, colon, or new sentence", a.File),
					})
				}
			}
			return out
		},
	})

	register(rule{
		id: "bolded-bullet-lead",
		doc: "a bullet may not open with a bolded term unless it is allowed " +
			"(params: allow)",
		eval: func(p params, c *ir.Change) []Finding {
			allowed := map[string]bool{}
			for _, a := range p.Strings("allow") {
				allowed[a] = true
			}
			var out []Finding
			for _, a := range allArtifacts(c) {
				for i, raw := range strings.Split(a.Text, "\n") {
					m := boldLeadRe.FindStringSubmatch(raw)
					if m == nil || allowed[m[1]] {
						continue
					}
					out = append(out, Finding{
						File: a.File, Line: i + 1,
						Msg: fmt.Sprintf("%s opens a bullet with the bolded term **%s**; use plain text or a sub-bullet",
							a.File, m[1]),
					})
				}
			}
			return out
		},
	})

	register(rule{
		id: "review-decision-current",
		doc: "a recorded review decision must exist and still describe the current " +
			"artifacts (params: accept, requireRecord)",
		eval: func(p params, c *ir.Change) []Finding {
			rec, err := review.LoadForChange(c)
			if err != nil {
				return []Finding{{File: review.RecordFile, Msg: err.Error()}}
			}
			accept := p.Strings("accept")
			if len(accept) == 0 {
				accept = []string{string(review.DecisionApproved)}
			}
			if rec == nil {
				// A repository that reviews only some changes sets requireRecord false
				// and still gets the staleness check on the ones it does review.
				if v, ok := p["requireRecord"].(bool); ok && !v {
					return nil
				}
				return []Finding{{
					File: review.RecordFile,
					Msg: fmt.Sprintf("no review decision recorded; run `specutil review set --change %s --decision %s`",
						c.Name, strings.Join(accept, "|")),
				}}
			}
			var out []Finding
			if !containsString(accept, string(rec.Decision)) {
				out = append(out, Finding{
					File: review.RecordFile,
					Msg: fmt.Sprintf("review decision is %q; the rubric accepts %s",
						rec.Decision, strings.Join(accept, ", ")),
				})
			}
			if cur := review.ChangeHash(c); cur != rec.ChangeHash {
				out = append(out, Finding{
					File: review.RecordFile,
					Msg: fmt.Sprintf("review decision is stale: the artifacts changed since it was recorded (reviewed %s, now %s)",
						rec.ChangeHash, cur),
				})
			}
			return out
		},
	})
}

// containsString reports whether list holds want.
func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// boldLeadRe matches a list bullet whose first token is bolded.
var boldLeadRe = regexp.MustCompile(`^\s*[-*]\s+\*\*([A-Za-z][A-Za-z0-9 _-]*)\*\*`)

// compile returns a compiled pattern, or nil when the pattern is empty or
// invalid. An invalid pattern disables its rule rather than aborting the run,
// and Resolve is where a malformed rubric should be caught.
func compile(pattern string) *regexp.Regexp {
	if pattern == "" {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}

// skipped reports whether a phase is exempt from a rule.
func skipped(skip *regexp.Regexp, ph ir.Phase) bool {
	return skip != nil && skip.MatchString(ph.Name)
}

// phaseLabel names a phase for a message, preferring its number and name.
func phaseLabel(ph ir.Phase) string {
	if ph.Number != "" {
		return ph.Number + ". " + ph.Name
	}
	return ph.Name
}

// cycleFrom returns the suffix of stack beginning at id, naming the cycle a
// back-edge to id closes.
func cycleFrom(stack []string, id string) []string {
	for i, s := range stack {
		if s == id {
			return append([]string(nil), stack[i:]...)
		}
	}
	return append([]string(nil), stack...)
}

// findLine returns the 1-based number of the first line satisfying match, or 0.
func findLine(text string, match func(string) bool) int {
	for i, line := range strings.Split(text, "\n") {
		if match(line) {
			return i + 1
		}
	}
	return 0
}
