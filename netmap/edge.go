// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package netmap

import (
	"context"
	"fmt"

	"github.com/caffix/stringset"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/quad"
)

// Constant values that represent the direction of edges during graph queries.
const (
	IN int = iota
	OUT
	BOTH
)

// Edge represents an edge in the graph.
type Edge struct {
	Predicate string
	From, To  Node
}

// UpsertEdge will create an edge in the database if it does not already exist.
func (g *Graph) UpsertEdge(ctx context.Context, edge *Edge) error {
	t := graph.NewTransaction()

	if edge.Predicate == "" {
		return fmt.Errorf("%s: UpsertEdge: Empty edge predicate", g.String())
	}

	from := g.NodeToID(edge.From)
	if from == "" {
		return fmt.Errorf("%s: UpsertEdge: Invalid 'from' node", g.String())
	}

	to := g.NodeToID(edge.To)
	if to == "" {
		return fmt.Errorf("%s: UpsertEdge: Invalid 'to' node", g.String())
	}

	if err := g.db.quadsUpsertEdge(t, edge.Predicate, from, to); err != nil {
		return err
	}
	return g.db.applyWithLock(t)
}

func (g *CayleyGraph) quadsUpsertEdge(t *graph.Transaction, predicate, from, to string) error {
	if predicate == "" {
		return fmt.Errorf("%s: quadsUpsertEdge: Empty edge predicate", g.String())
	}

	if from == "" {
		return fmt.Errorf("%s: quadsUpsertEdge: Invalid from node", g.String())
	}

	if to == "" {
		return fmt.Errorf("%s: quadsUpsertEdge: Invalid to node", g.String())
	}

	t.AddQuad(quad.MakeIRI(from, predicate, to, ""))
	return nil
}

// ReadEdges implements the GraphDatabase interface.
func (g *Graph) ReadEdges(ctx context.Context, node Node, predicates ...string) ([]*Edge, error) {
	var edges []*Edge

	if e, err := g.ReadInEdges(ctx, node, predicates...); err == nil {
		edges = append(edges, e...)
	}

	if e, err := g.ReadOutEdges(ctx, node, predicates...); err == nil {
		edges = append(edges, e...)
	}

	if len(edges) == 0 {
		return nil, fmt.Errorf("%s: ReadEdges: Failed to discover edges for the node %s", g.String(), g.NodeToID(node))
	}

	return edges, nil
}

// CountEdges counts the total number of edges to a node.
func (g *Graph) CountEdges(ctx context.Context, node Node, predicates ...string) (int, error) {
	var count int

	if c, err := g.CountInEdges(ctx, node, predicates...); err == nil {
		count += c
	} else {
		return 0, fmt.Errorf("%s: CountEdges: %v", g.String(), err)
	}

	if c, err := g.CountOutEdges(ctx, node, predicates...); err == nil {
		count += c
	}

	return count, nil
}

// ReadInEdges implements the GraphDatabase interface.
func (g *Graph) ReadInEdges(ctx context.Context, node Node, predicates ...string) ([]*Edge, error) {
	g.db.Lock()
	defer g.db.Unlock()

	nstr := g.NodeToID(node)
	if nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return nil, fmt.Errorf("%s: ReadInEdges: Invalid node reference argument", g.String())
	}

	var preds []interface{}
	filter := stringset.New()
	defer filter.Close()

	for _, pred := range predicates {
		if !filter.Has(pred) {
			filter.Insert(pred)
			preds = append(preds, quad.IRI(pred))
		}
	}

	p := cayley.StartPath(g.db.store, quad.IRI(nstr))
	if len(predicates) == 0 {
		p = p.InWithTags([]string{"predicate"})
	} else {
		p = p.InWithTags([]string{"predicate"}, preds...)
	}
	p = p.Has(quad.IRI("type")).Tag("object")

	var edges []*Edge
	err := p.Iterate(ctx).TagValues(nil, func(m map[string]quad.Value) error {
		edges = append(edges, &Edge{
			Predicate: valToStr(m["predicate"]),
			From:      valToStr(m["object"]),
			To:        node,
		})
		return nil
	})

	if err == nil && len(edges) == 0 {
		return nil, fmt.Errorf("%s: ReadInEdges: Failed to discover edges coming into the node %s", g.String(), nstr)
	}
	return edges, err
}

// CountInEdges implements the GraphDatabase interface.
func (g *Graph) CountInEdges(ctx context.Context, node Node, predicates ...string) (int, error) {
	g.db.Lock()
	defer g.db.Unlock()

	nstr := g.NodeToID(node)
	if nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return 0, fmt.Errorf("%s: CountInEdges: Invalid node reference argument", g.String())
	}

	p := cayley.StartPath(g.db.store, quad.IRI(nstr))
	if len(predicates) == 0 {
		p = p.In()
	} else {
		p = p.In(strsToVals(predicates...))
	}
	p = p.Has(quad.IRI("type"))
	count, err := p.Iterate(ctx).Count()

	return int(count), err
}

// ReadOutEdges implements the GraphDatabase interface.
func (g *Graph) ReadOutEdges(ctx context.Context, node Node, predicates ...string) ([]*Edge, error) {
	g.db.Lock()
	defer g.db.Unlock()

	nstr := g.NodeToID(node)
	if nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return nil, fmt.Errorf("%s: ReadOutEdges: Invalid node reference argument", g.String())
	}

	var preds []interface{}
	filter := stringset.New()
	defer filter.Close()

	for _, pred := range predicates {
		if !filter.Has(pred) {
			filter.Insert(pred)
			preds = append(preds, quad.IRI(pred))
		}
	}

	p := cayley.StartPath(g.db.store, quad.IRI(nstr))
	if len(predicates) == 0 {
		p = p.OutWithTags([]string{"predicate"})
	} else {
		p = p.OutWithTags([]string{"predicate"}, preds...)
	}
	p = p.Has(quad.IRI("type")).Tag("object")

	var edges []*Edge
	err := p.Iterate(ctx).TagValues(nil, func(m map[string]quad.Value) error {
		edges = append(edges, &Edge{
			Predicate: valToStr(m["predicate"]),
			From:      node,
			To:        valToStr(m["object"]),
		})
		return nil
	})

	if err == nil && len(edges) == 0 {
		return nil, fmt.Errorf("%s: ReadOutEdges: Failed to discover edges leaving the node %s", g.String(), nstr)
	}
	return edges, err
}

// CountOutEdges implements the GraphDatabase interface.
func (g *Graph) CountOutEdges(ctx context.Context, node Node, predicates ...string) (int, error) {
	g.db.Lock()
	defer g.db.Unlock()

	nstr := g.NodeToID(node)
	if nstr == "" || !g.db.nodeExists(ctx, nstr, "") {
		return 0, fmt.Errorf("%s: CountOutEdges: Invalid node reference argument", g.String())
	}

	p := cayley.StartPath(g.db.store, quad.IRI(nstr))
	if len(predicates) == 0 {
		p = p.Out()
	} else {
		p = p.Out(strsToVals(predicates...))
	}
	p = p.Has(quad.IRI("type"))
	count, err := p.Iterate(ctx).Count()

	return int(count), err
}

// DeleteEdge implements the GraphDatabase interface.
func (g *Graph) DeleteEdge(ctx context.Context, edge *Edge) error {
	g.db.Lock()
	defer g.db.Unlock()

	from := g.NodeToID(edge.From)
	to := g.NodeToID(edge.To)
	if edge.Predicate == "" || from == "" || to == "" {
		return fmt.Errorf("%s: DeleteEdge: Invalid edge reference argument", g.String())
	}

	return g.db.store.RemoveQuad(quad.MakeIRI(from, edge.Predicate, to, ""))
}
