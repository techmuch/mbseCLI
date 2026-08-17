package parser

import "testing"

func TestParse_BasicHierarchy(t *testing.T) {
	src := []byte(`
package Drone {
    part def Airframe {
        part battery : Battery;
        in port powerIn : ElectricalPort;
    }
    part def Battery {
        attribute capacity : Real;
    }
}
`)
	g := Parse("drone.sysml", src)
	if g.Root == nil {
		t.Fatal("expected non-nil root")
	}

	el, ok := g.ByFQN["Drone::Airframe"]
	if !ok {
		t.Fatal("expected Drone::Airframe in index")
	}
	if el.Kind != "part def" {
		t.Errorf("kind = %q, want %q", el.Kind, "part def")
	}

	battery, ok := g.ByFQN["Drone::Airframe::battery"]
	if !ok {
		t.Fatal("expected Drone::Airframe::battery")
	}
	if battery.Type != "Battery" {
		t.Errorf("type = %q, want Battery", battery.Type)
	}

	port, ok := g.ByFQN["Drone::Airframe::powerIn"]
	if !ok {
		t.Fatal("expected port Drone::Airframe::powerIn")
	}
	if port.Kind != "port" {
		t.Errorf("kind = %q, want port", port.Kind)
	}

	attr, ok := g.ByFQN["Drone::Battery::capacity"]
	if !ok {
		t.Fatal("expected attribute Drone::Battery::capacity")
	}
	if attr.Kind != "attribute" {
		t.Errorf("kind = %q, want attribute", attr.Kind)
	}
}

func TestParse_Doc(t *testing.T) {
	// Closing brace must be on its own line — this tokenizer is line-based
	// and doesn't detect a block opened and closed on the same line.
	src := []byte(`
package P {
    requirement def R1 {
        doc /* must be light */
    }
}
`)
	g := Parse("p.sysml", src)
	r, ok := g.ByFQN["P::R1"]
	if !ok {
		t.Fatal("expected P::R1")
	}
	if r.Kind != "requirement" {
		t.Errorf("kind = %q, want requirement", r.Kind)
	}
	if r.Doc != "must be light" {
		t.Errorf("doc = %q, want %q", r.Doc, "must be light")
	}
}

func TestParse_EmptyInput(t *testing.T) {
	g := Parse("empty.sysml", []byte(""))
	if g.Root == nil {
		t.Fatal("expected synthetic root for empty input")
	}
	if len(g.Root.Children) != 0 {
		t.Errorf("expected no children, got %d", len(g.Root.Children))
	}
}

func TestParse_UnnamedDeclarationGetsSyntheticName(t *testing.T) {
	// Real SysML v2 allows anonymous usages. The tokenizer should still
	// produce a stable, non-empty name/FQN rather than dropping the node.
	src := []byte(`
package P {
    part : Widget;
}
`)
	g := Parse("p.sysml", src)
	part := g.Root.Children[0]
	if len(part.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(part.Children))
	}
	if part.Children[0].Name == "" {
		t.Error("expected a synthetic name for the anonymous part")
	}
}
