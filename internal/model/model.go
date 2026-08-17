// Package model defines the in-memory intermediate representation (IR) that
// the parser produces and the server/UI consume. It is intentionally decoupled
// from the SysML v2 grammar so the parser can be swapped for a real ANTLR-based
// implementation later without changing downstream consumers.
package model

// Element is a single node in a parsed SysML v2 model (a package, part def,
// part, port, action, state, requirement, attribute, etc). Elements nest to
// form a containment tree; FQN is the stable identifier used to anchor
// feedback notes and to correlate elements across live-reloads.
type Element struct {
	FQN      string     `json:"fqn"`
	Name     string     `json:"name"`
	Kind     string     `json:"kind"` // package, part def, part, port, action, state, requirement, attribute, item, connection, ...
	Type     string     `json:"type,omitempty"`
	Doc      string     `json:"doc,omitempty"`
	Line     int        `json:"line"`
	Children []*Element `json:"children,omitempty"`
}

// Graph is a full parsed model: the containment tree plus a flat FQN index
// for O(1) lookups (tree navigation, feedback anchoring, selection sync).
type Graph struct {
	SourceFile string              `json:"sourceFile"`
	Root       *Element            `json:"root"`
	ByFQN      map[string]*Element `json:"-"`
	Errors     []ParseError        `json:"errors,omitempty"`
}

// ParseError describes a recoverable parse issue. The parser is fault-tolerant
// by design: a syntax error should degrade the affected subtree, not blank
// out the whole UI while the user is mid-edit in their own editor.
type ParseError struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// Index walks the tree and (re)builds ByFQN. Call after parsing or mutating
// the tree.
func (g *Graph) Index() {
	g.ByFQN = make(map[string]*Element)
	if g.Root == nil {
		return
	}
	var walk func(*Element)
	walk = func(e *Element) {
		g.ByFQN[e.FQN] = e
		for _, c := range e.Children {
			walk(c)
		}
	}
	walk(g.Root)
}
