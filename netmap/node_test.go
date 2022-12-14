// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package netmap

import (
	"context"
	"testing"

	"github.com/caffix/stringset"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/quad"
)

func TestNodeToID(t *testing.T) {
	id := "test"
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	if node := Node(id); g.NodeToID(node) != id {
		t.Errorf("The graph node id was not properly returned by NodeToID")
	}
}

func TestAllNodesOfType(t *testing.T) {
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	// setup the data in the graph
	if err := g.db.store.AddQuad(quad.Make(quad.IRI("test"), quad.IRI("type"), quad.String("test"), nil)); err != nil {
		t.Errorf("Failed to add the quad: %v", err)
	}

	if nodes, err := g.AllNodesOfType(context.Background(), "test"); err != nil {
		t.Errorf("AllNodesOfType returned an error for a non-empty graph and matching constraints")
	} else if len(nodes) == 0 {
		t.Errorf("AllNodesOfType returned an empty slice of nodes for a non-empty graph and matching constraints")
	}

	if nodes, err := g.AllNodesOfType(context.Background(), "do_not_match"); err == nil {
		t.Errorf("AllNodesOfType returned no error for a non-empty graph and differing constraints")
	} else if len(nodes) > 0 {
		t.Errorf("AllNodesOfType returned non-empty slice of nodes for a non-empty graph and differing constraints")
	}
}

func TestAllOutNodes(t *testing.T) {
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	vBob := quad.IRI("Bob")
	vAlice := quad.IRI("Alice")
	vCharles := quad.IRI("Charles")
	knows := quad.IRI("knows")
	vType := quad.IRI("type")

	if nodes, err := g.AllOutNodes(context.Background(), "Bob"); err == nil {
		t.Errorf("AllOutNodes returned no error for an empty graph")
	} else if len(nodes) > 0 {
		t.Errorf("AllOutNodes returned a non-empty slice of nodes on an empty graph")
	}

	// setup the initial data in the graph
	if err := g.db.store.AddQuad(quad.Make(vBob, knows, vAlice, nil)); err != nil {
		t.Errorf("Failed to add the bob know alice quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vBob, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the bob quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vAlice, knows, vCharles, nil)); err != nil {
		t.Errorf("Failed to add the alice knows charles quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vAlice, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the alice quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vCharles, knows, vAlice, nil)); err != nil {
		t.Errorf("Failed to add the charles knows alice quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vCharles, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the charles quad: %v", err)
	}

	if nodes, err := g.AllOutNodes(context.Background(), "Bob"); err != nil {
		t.Errorf("AllOutNodes returned an error when out nodes existed from the node")
	} else if len(nodes) != 1 {
		t.Errorf("AllOutNodes returned the incorrent number of nodes in the slice")
	} else if g.NodeToID(nodes[0]) != "Alice" {
		t.Errorf("AllOutNodes returned a slice with the wrong node")
	}

	if err := g.db.store.AddQuad(quad.Make(vBob, knows, vCharles, nil)); err != nil {
		t.Errorf("Failed to add the bob knows charles quad: %v", err)
	}

	nodes, err := g.AllOutNodes(context.Background(), "Bob")
	if err != nil {
		t.Errorf("AllOutNodes returned an error when out nodes existed from the node")
	} else if len(nodes) != 2 {
		t.Errorf("AllOutNodes returned the incorrent number of nodes in the slice")
	}

	got := stringset.New()
	defer got.Close()

	expected := stringset.New()
	defer expected.Close()

	expected.InsertMany("Alice", "Charles")
	for _, node := range nodes {
		got.Insert(g.NodeToID(node))
	}
	expected.Subtract(got)
	if expected.Len() != 0 {
		t.Errorf("AllOutNodes returned a slice with the wrong nodes: %v", got.Slice())
	}
}

func TestUpsertNode(t *testing.T) {
	name := "test"
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	ctx := context.Background()
	if _, err := g.UpsertNode(ctx, "", name); err == nil {
		t.Errorf("UpsertNode did not return an error when the id is invalid")
	}

	if _, err := g.UpsertNode(ctx, name, ""); err == nil {
		t.Errorf("UpsertNode did not return an error when the type is invalid")
	}

	if node, err := g.UpsertNode(ctx, name, name); err != nil {
		t.Errorf("UpsertNode returned an error when the arguments are valid")
	} else if g.NodeToID(node) != name {
		t.Errorf("UpsertNode did not return the node with the correct identifier")
	}
	// Try to insert the same node again
	if node, err := g.UpsertNode(ctx, name, name); err != nil {
		t.Errorf("UpsertNode returned an error on a second execution with the same valid arguments")
	} else if g.NodeToID(node) != name {
		t.Errorf("UpsertNode did not return the node with the correct identifier on a second execution with the same valid arguments")
	}

	// Check if the node was properly entered into the graph database
	p := cayley.StartPath(g.db.store, quad.IRI(name)).Has(quad.IRI("type"), quad.String(name))
	if first, err := p.Iterate(ctx).FirstValue(nil); err != nil || valToStr(first) != "test" {
		t.Errorf("UpsertNode failed to enter the node: expected %s and got %s", name, valToStr(first))
	}
}

func TestReadNode(t *testing.T) {
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	bob := "Bob"
	bType := "Person"
	vBob := quad.IRI(bob)
	vType := quad.IRI("type")

	if _, err := g.ReadNode(context.Background(), "", bType); err == nil {
		t.Errorf("ReadNode returned no error when given an invalid id argument")
	}
	if _, err := g.ReadNode(context.Background(), bob, ""); err == nil {
		t.Errorf("ReadNode returned no error when given an invalid type argument")
	}
	if _, err := g.ReadNode(context.Background(), bob, bType); err == nil {
		t.Errorf("ReadNode returned no error when given arguments for a non-existent node")
	}

	// setup the initial data in the graph
	if err := g.db.store.AddQuad(quad.Make(vBob, vType, bType, nil)); err != nil {
		t.Errorf("Failed to add the bob quad: %v", err)
	}

	if node, err := g.ReadNode(context.Background(), bob, bType); err != nil {
		t.Errorf("ReadNode returned an error when given valid arguments")
	} else if g.NodeToID(node) != bob {
		t.Errorf("ReadNode returned a node that does not match the arguments")
	}
}

func TestDeleteNode(t *testing.T) {
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	if err := g.DeleteNode(context.Background(), ""); err == nil {
		t.Errorf("DeleteNode returned no error when provided an invalid argument")
	}

	vBob := quad.IRI("Bob")
	vAlice := quad.IRI("Alice")
	vCharles := quad.IRI("Charles")
	knows := quad.IRI("knows")
	likes := quad.IRI("likes")
	vType := quad.IRI("type")

	if err := g.DeleteNode(context.Background(), "Bob"); err == nil {
		t.Errorf("DeleteNode returned no error when the argument node did not exist")
	}
	// setup the initial data in the graph
	if err := g.db.store.AddQuad(quad.Make(vBob, knows, vAlice, nil)); err != nil {
		t.Errorf("Failed to add the bob knows alice quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vBob, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the bob quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vBob, knows, vCharles, nil)); err != nil {
		t.Errorf("Failed to add the bob knows charles quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vCharles, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the charles quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vBob, likes, "Go", nil)); err != nil {
		t.Errorf("Failed to add the bob likes Go quad: %v", err)
	}
	if err := g.db.store.AddQuad(quad.Make(vBob, likes, "Automation", nil)); err != nil {
		t.Errorf("Failed to add the bob likes Automation quad: %v", err)
	}

	if err := g.DeleteNode(context.Background(), "Bob"); err != nil {
		t.Errorf("DeleteNode returned an error when provided a valid node: %v", err)
	}
	// Check that no quads with 'Bob' as a subject exist
	p := cayley.StartPath(g.db.store, vBob).Out()
	if count, err := p.Iterate(context.Background()).Count(); err == nil && count != 0 {
		t.Errorf("DeleteNode did not remove all the quads with 'Bob' as the subject")
	}
}

func TestWriteNodeQuads(t *testing.T) {
	cay := NewCayleyGraphMemory()
	g := NewGraph(cay)
	defer g.Close()

	vBob := quad.IRI("Bob")
	vAlice := quad.IRI("Alice")
	vCharles := quad.IRI("Charles")
	knows := quad.IRI("knows")
	vType := quad.IRI("type")
	// setup the initial data in the graph
	expected := stringset.New()
	defer expected.Close()

	if err := g.db.store.AddQuad(quad.Make(vBob, knows, vAlice, nil)); err != nil {
		t.Errorf("Failed to add the bob knows alice quad: %v", err)
	}
	expected.Insert("BobknowsAlice")
	if err := g.db.store.AddQuad(quad.Make(vBob, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the bob quad: %v", err)
	}
	expected.Insert("BobtypePerson")
	if err := g.db.store.AddQuad(quad.Make(vAlice, knows, vCharles, nil)); err != nil {
		t.Errorf("Failed to add the alice knows charles quad: %v", err)
	}
	expected.Insert("AliceknowsCharles")
	if err := g.db.store.AddQuad(quad.Make(vAlice, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the alice quad: %v", err)
	}
	expected.Insert("AlicetypePerson")
	if err := g.db.store.AddQuad(quad.Make(vCharles, knows, vAlice, nil)); err != nil {
		t.Errorf("Failed to add the charles knows alice quad: %v", err)
	}
	expected.Insert("CharlesknowsAlice")
	if err := g.db.store.AddQuad(quad.Make(vCharles, vType, "Person", nil)); err != nil {
		t.Errorf("Failed to add the charles quad: %v", err)
	}
	expected.Insert("CharlestypePerson")

	dup := NewGraph(NewCayleyGraphMemory())
	defer dup.Close()

	nodes, _ := g.AllNodesOfType(context.Background(), "Person")
	if err := dup.WriteNodeQuads(context.Background(), g, nodes); err != nil {
		t.Errorf("WriteNodeQuads returned an error when provided valid arguments")
	}

	got := stringset.New()
	defer got.Close()

	p := cayley.StartPath(dup.db.store).Tag("subject").OutWithTags([]string{"predicate"}).Tag("object")
	err := p.Iterate(context.Background()).TagValues(nil, func(m map[string]quad.Value) error {
		sub := valToStr(m["subject"])
		pred := valToStr(m["predicate"])
		obj := valToStr(m["object"])
		got.Insert(sub + pred + obj)
		return nil
	})
	if err != nil {
		t.Errorf("Failed to iterate over the tags: %v", err)
	}

	expected.Subtract(got)
	if expected.Len() != 0 {
		t.Errorf("WriteNodeQuads did not replicate all the quads")
	}
}
