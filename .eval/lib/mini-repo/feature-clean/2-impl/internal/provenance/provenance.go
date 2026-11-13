// Package provenance builds the source metadata attached to generated
// artifacts.
package provenance

import (
	"fmt"

	"github.com/unbound-force/unbound-force/internal/refs"
)

// Record identifies the source an artifact was produced from: the ref
// as the caller named it, and the commit that ref pointed at when the
// artifact was stamped. Ref names move; Commit does not.
type Record struct {
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

// Stamp returns the provenance record for an artifact built from ref
// in the repository at dir. It fails rather than return a Record
// without a commit, so no artifact carries unpinned provenance.
func Stamp(dir, ref string) (Record, error) {
	commit, err := refs.Lookup(dir, ref)
	if err != nil {
		return Record{}, fmt.Errorf("stamp provenance: %w", err)
	}
	return Record{Ref: refs.Normalize(ref), Commit: commit}, nil
}
