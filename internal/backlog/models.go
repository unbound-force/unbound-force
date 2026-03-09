package backlog

import "time"

// Item represents a backlog item parsed from a markdown file with YAML frontmatter.
type Item struct {
	ID                string    `yaml:"id"`
	Title             string    `yaml:"title"`
	Type              string    `yaml:"type"`
	Priority          string    `yaml:"priority"`
	Status            string    `yaml:"status"`
	Sprint            string    `yaml:"sprint,omitempty"`
	EffortEstimate    string    `yaml:"effort_estimate,omitempty"`
	Dependencies      []string  `yaml:"dependencies,omitempty"`
	RelatedSpecs      []string  `yaml:"related_specs,omitempty"`
	GitHubIssueNumber *int      `yaml:"github_issue_number,omitempty"`
	CreatedAt         time.Time `yaml:"created_at"`
	ModifiedAt        time.Time `yaml:"modified_at"`
	Body              string    `yaml:"-"` // Stored separately from frontmatter
}
