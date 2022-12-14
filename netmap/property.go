// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package netmap

import (
	"context"
	"fmt"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/quad"
)

// Property represents a node property.
type Property struct {
	Predicate string
	Value     quad.Value
}

// UpsertProperty implements the GraphDatabase interface.
func (g *Graph) UpsertProperty(ctx context.Context, node Node, predicate, value string) error {
	g.db.Lock()
	defer g.db.Unlock()

	id := g.NodeToID(node)
	if !g.db.nodeExists(ctx, id, "") {
		return fmt.Errorf("%s: UpsertProperty: Invalid node reference argument", g.String())
	}

	t := graph.NewTransaction()
	if err := g.db.quadsUpsertProperty(t, id, predicate, value); err != nil {
		return err
	}
	return g.db.store.ApplyTransaction(t)
}

func (g *CayleyGraph) quadsUpsertProperty(t *graph.Transaction, id, predicate string, value interface{}) error {
	if id == "" {
		return fmt.Errorf("%s: quadsUpsertProperty: Invalid node reference argument", g.String())
	} else if predicate == "" {
		return fmt.Errorf("%s: quadsUpsertProperty: Empty predicate argument", g.String())
	}

	t.AddQuad(quad.Make(quad.IRI(id), quad.IRI(predicate), value, nil))
	return nil
}

// ReadProperties implements the GraphDatabase interface.
func (g *Graph) ReadProperties(ctx context.Context, node Node, predicates ...string) ([]*Property, error) {
	g.db.Lock()
	defer g.db.Unlock()

	nstr := g.NodeToID(node)
	var properties []*Property

	if nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return properties, fmt.Errorf("%s: ReadProperties: Invalid node reference argument", g.String())
	}

	p := cayley.StartPath(g.db.store, quad.IRI(nstr))
	if len(predicates) == 0 {
		p = p.OutWithTags([]string{"predicate"})
	} else {
		p = p.OutWithTags([]string{"predicate"}, strsToVals(predicates...))
	}
	p = p.Tag("object")

	err := p.Iterate(ctx).TagValues(nil, func(m map[string]quad.Value) error {
		// Check if this is actually a node and not a property
		if !isIRI(m["object"]) {
			properties = append(properties, &Property{
				Predicate: valToStr(m["predicate"]),
				Value:     m["object"],
			})
		}
		return nil
	})
	// Given the data model, valid nodes should always have at least one
	// property, and for that reason, it doesn't need to be checked here
	return properties, err
}

// CountProperties implements the GraphDatabase interface.
func (g *Graph) CountProperties(ctx context.Context, node Node, predicates ...string) (int, error) {
	g.db.Lock()
	defer g.db.Unlock()

	nstr := g.NodeToID(node)
	if nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return 0, fmt.Errorf("%s: CountProperties: Invalid node reference argument", g.String())
	}

	p := cayley.StartPath(g.db.store, quad.IRI(nstr))
	if len(predicates) == 0 {
		p = p.Out()
	} else {
		p = p.Out(strsToVals(predicates...))
	}

	var count int
	err := p.Iterate(ctx).EachValue(nil, func(value quad.Value) error {
		if !isIRI(value) {
			count++
		}
		return nil
	})
	return count, err
}

// DeleteProperty implements the GraphDatabase interface.
func (g *Graph) DeleteProperty(ctx context.Context, node Node, predicate string, value interface{}) error {
	g.db.Lock()
	defer g.db.Unlock()

	v, ok := quad.AsValue(value)
	nstr := g.NodeToID(node)
	if !ok || nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return fmt.Errorf("%s: DeleteProperty: Invalid node reference argument", g.String())
	}

	return g.db.store.RemoveQuad(quad.Make(quad.IRI(nstr), quad.IRI(predicate), v, nil))
}
