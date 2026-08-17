// Package parser turns SysML v2 textual notation (.sysml) source into a
// model.Graph.
//
// NOTE: This is a deliberately naive, line-oriented tokenizer covering the
// common declaration forms (package, part def, part, port, action, state,
// requirement, attribute, item) so the rest of the stack (watcher, server,
// UI) has a real end-to-end path to build against. It is NOT a conformant
// SysML v2 / KerML parser. Swap this out for a generated ANTLR4 parser
// (SysML.g4 / KerML.g4, `antlr4 -Dlanguage=Go`) once the visualization
// pipeline is validated — model.Graph is the seam to preserve.
package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"mbsecli/internal/model"
)

// declRe matches lines like:
//
//	part def Vehicle {
//	part battery : Battery {
//	port in p1 : ElectricalPort;
//	action Fly {
//	state def FlightMode {
//	requirement def R1 { doc /* ... */
//	attribute mass : Real;
//	package Drone {
var declRe = regexp.MustCompile(
	`^\s*(package|part def|part|port|in port|out port|inout port|action def|action|state def|state|requirement def|requirement|attribute|item def|item|connection|interface def|interface)\s+([A-Za-z_][\w]*)?\s*(?::\s*([A-Za-z_][\w:]*))?\s*([{;])?`,
)

var docRe = regexp.MustCompile(`/\*(.*?)\*/`)

// Parse reads SysML v2-ish source and returns a best-effort model.Graph.
// It never returns a nil Root: on total failure it returns a synthetic root
// so the UI has something stable to render.
func Parse(sourceFile string, src []byte) *model.Graph {
	root := &model.Element{FQN: "", Name: rootName(sourceFile), Kind: "package"}
	g := &model.Graph{SourceFile: sourceFile, Root: root}

	stack := []*model.Element{root}
	depth := 0

	scanner := bufio.NewScanner(strings.NewReader(string(src)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Closing brace pops the containment stack.
		if trimmed == "}" || strings.HasPrefix(trimmed, "}") {
			if depth > 0 {
				stack = stack[:len(stack)-1]
				depth--
			}
			continue
		}

		if m := declRe.FindStringSubmatch(trimmed); m != nil {
			kind, name, typ, terminator := normalizeKind(m[1]), m[2], m[3], m[4]
			if name == "" {
				name = fmt.Sprintf("%s_%d", sanitize(kind), lineNo)
			}
			parent := stack[len(stack)-1]
			fqn := name
			if parent.FQN != "" {
				fqn = parent.FQN + "::" + name
			}
			el := &model.Element{
				FQN:  fqn,
				Name: name,
				Kind: kind,
				Type: typ,
				Line: lineNo,
			}
			if doc := docRe.FindStringSubmatch(trimmed); doc != nil {
				el.Doc = strings.TrimSpace(doc[1])
			}
			parent.Children = append(parent.Children, el)

			if terminator == "{" {
				stack = append(stack, el)
				depth++
			}
			continue
		}

		// A `doc /* ... */` statement on its own line — the idiomatic form,
		// since it's normally the first statement in a definition's body.
		// Attached to the immediately enclosing block; a real grammar-based
		// parser would instead attribute it per SysML v2's actual doc-usage
		// rules (which allow docs on nested elements too).
		if depth > 0 && strings.HasPrefix(trimmed, "doc") {
			if m := docRe.FindStringSubmatch(trimmed); m != nil {
				stack[len(stack)-1].Doc = strings.TrimSpace(m[1])
			}
			continue
		}

		// Unrecognized line inside a block: not fatal, just note it so the
		// UI can surface "partial parse" state without losing the rest of
		// the tree.
		if depth > 0 && !strings.HasSuffix(trimmed, ";") {
			// Likely a continuation line (multi-line doc, long type list, etc).
			// Skip silently — this is exactly the kind of thing a real
			// grammar-based parser will handle correctly.
			continue
		}
	}

	g.Index()
	return g
}

func normalizeKind(raw string) string {
	switch raw {
	case "in port", "out port", "inout port":
		return "port"
	case "action def":
		return "action"
	case "state def":
		return "state"
	case "requirement def":
		return "requirement"
	case "item def":
		return "item"
	case "interface def":
		return "interface"
	default:
		return raw
	}
}

func sanitize(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

func rootName(sourceFile string) string {
	base := sourceFile
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	return strings.TrimSuffix(base, ".sysml")
}
