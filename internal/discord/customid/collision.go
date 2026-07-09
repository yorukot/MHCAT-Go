package customid

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Rule struct {
	Name     string   `json:"name"`
	Pattern  string   `json:"pattern"`
	Examples []string `json:"examples,omitempty"`
	Broad    bool     `json:"broad,omitempty"`
}

type CollisionIssue struct {
	Left   string `json:"left"`
	Right  string `json:"right"`
	Reason string `json:"reason"`
}

type CollisionReport struct {
	Issues []CollisionIssue `json:"issues"`
}

func AnalyzeRules(rules []Rule) CollisionReport {
	var issues []CollisionIssue
	for i := range rules {
		for j := i + 1; j < len(rules); j++ {
			if reason := collisionReason(rules[i], rules[j]); reason != "" {
				issues = append(issues, CollisionIssue{Left: rules[i].Name, Right: rules[j].Name, Reason: reason})
			}
		}
	}
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Left != issues[j].Left {
			return issues[i].Left < issues[j].Left
		}
		if issues[i].Right != issues[j].Right {
			return issues[i].Right < issues[j].Right
		}
		return issues[i].Reason < issues[j].Reason
	})
	return CollisionReport{Issues: issues}
}

func FormatCollisionReport(w io.Writer, report CollisionReport, format string) error {
	normalized := CollisionReport{Issues: append([]CollisionIssue(nil), report.Issues...)}
	sort.SliceStable(normalized.Issues, func(i, j int) bool {
		if normalized.Issues[i].Left != normalized.Issues[j].Left {
			return normalized.Issues[i].Left < normalized.Issues[j].Left
		}
		return normalized.Issues[i].Right < normalized.Issues[j].Right
	})
	if format == "json" {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(normalized)
	}
	for _, issue := range normalized.Issues {
		if _, err := fmt.Fprintf(w, "collision left=%s right=%s reason=%q\n", issue.Left, issue.Right, issue.Reason); err != nil {
			return err
		}
	}
	return nil
}

func collisionReason(left, right Rule) string {
	if left.Pattern == right.Pattern {
		return "duplicate pattern"
	}
	if left.Broad || right.Broad {
		return "includes-style broad rule"
	}
	if strings.HasPrefix(left.Pattern, right.Pattern) || strings.HasPrefix(right.Pattern, left.Pattern) {
		return "prefix overlap"
	}
	if !strings.Contains(left.Pattern, ":") && !strings.Contains(right.Pattern, ":") {
		if strings.Contains(left.Pattern, right.Pattern) || strings.Contains(right.Pattern, left.Pattern) {
			return "delimiterless overlap"
		}
	}
	return ""
}

func DocumentedLegacyRules() []Rule {
	return []Rule{
		{Name: "poll-vote", Pattern: "poll_*"},
		{Name: "role-add", Pattern: "*add", Broad: true},
		{Name: "role-delete", Pattern: "*delete", Broad: true},
		{Name: "rank-text", Pattern: "*text_rank", Broad: true},
		{Name: "verification", Pattern: "*verification"},
	}
}
