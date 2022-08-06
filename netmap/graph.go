// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package netmap

import (
	"context"
	"sync"
	"time"

	"github.com/cayleygraph/quad"
)

// Graph implements the network infrastructure data model.
type Graph struct {
	db            *CayleyGraph
	alreadyClosed bool

	// eventFinishes maintains a cache of the latest finish time for each event
	// This reduces roundtrips to the graph when adding nodes to events.
	eventFinishes   map[string]time.Time
	eventFinishLock sync.Mutex
}

// NewGraph accepts a graph database that stores the Graph created and maintained by the data model.
func NewGraph(database *CayleyGraph) *Graph {
	if database == nil {
		return nil
	}

	return &Graph{
		db:            database,
		eventFinishes: make(map[string]time.Time),
	}
}

// Close will close the graph database being used by the Graph receiver.
func (g *Graph) Close() {
	if !g.alreadyClosed {
		g.alreadyClosed = true
		g.db.Close()
	}
}

// String returns the name of the graph database used by the Graph.
func (g *Graph) String() string {
	return g.db.String()
}

// DumpGraph prints all data currently in the graph.
func (g *Graph) DumpGraph(ctx context.Context) string {
	return g.db.DumpGraph(ctx)
}

func isIRI(val quad.Value) bool {
	_, ok := val.(quad.IRI)

	return ok
}
