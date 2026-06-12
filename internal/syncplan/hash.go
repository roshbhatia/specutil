// Package syncplan implements the deterministic, offline half of the sync
// story: stable item identity, the per-change lockfile (lock get/set), the
// create/update/orphan plan, and drift diffing. No network I/O happens here;
// the shipped skills drive the agent to apply a plan via MCP and record results
// back through lock set.
package syncplan

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// emphasisRe strips markdown emphasis/code markers so they do not perturb the
// normalized identity key.
var emphasisRe = regexp.MustCompile("[*_`]")

// wsRe collapses internal whitespace runs to a single space.
var wsRe = regexp.MustCompile(`\s+`)

// trailingPunctRe trims trailing sentence punctuation that minor edits add or
// drop without changing meaning.
var trailingPunctRe = regexp.MustCompile(`[.,;:!?\s]+$`)

// normalize produces the position-independent, edit-tolerant key used for
// identity: lowercased, emphasis-stripped, whitespace-collapsed, trailing
// punctuation removed. It deliberately discards leading task numbers (the
// caller passes already number-free text) so renumbering preserves identity.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = emphasisRe.ReplaceAllString(s, "")
	s = wsRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = trailingPunctRe.ReplaceAllString(s, "")
	return s
}

// Identity is the stable lock key for an item. It is built from the normalized
// phase name and normalized item text, so it survives task renumbering and
// minor text edits (which the normalization absorbs) while still distinguishing
// genuinely different items. The phase name disambiguates identically-worded
// tasks living in different phases.
func Identity(phaseName, text string) string {
	key := normalize(phaseName) + "\n" + normalize(text)
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])[:16]
}

// ContentHash is the exact-content fingerprint stored alongside the external ID
// so plan can tell an unchanged item (skip) from an edited one (update). Unlike
// Identity it is NOT normalized: any byte change flips it.
func ContentHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])[:16]
}
