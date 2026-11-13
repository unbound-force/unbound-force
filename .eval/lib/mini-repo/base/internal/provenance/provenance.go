// Package provenance builds the source metadata attached to generated
// artifacts.
package provenance

import "github.com/unbound-force/unbound-force/internal/refs"

// Record identifies the source an artifact was produced from.
type Record struct {
	Ref string `json:"ref"`
}

// Stamp returns the provenance record for an artifact built from ref.
func Stamp(ref string) Record {
	return Record{Ref: refs.Normalize(ref)}
}
