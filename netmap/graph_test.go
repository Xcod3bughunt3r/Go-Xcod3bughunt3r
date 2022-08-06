// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package netmap

import (
	"testing"
)

var graphTest = []struct {
	Addr      string
	Source    string
	Tag       string
	FQDN      string
	EventID   string
	Name      string
	ASN       int
	ASNString string
	CIDR      string
	Desc      string
	Service   string
	ID        string
	Domain    string
}{
	{
		"testaddr",
		"testsource",
		"testtag",
		"www.owasp.org",
		"ef9f9475-34eb-465e-81eb-77c944822d0f",
		"testname",
		667,
		"667",
		"10.0.0.0/8",
		"a test description",
		"testservice.com",
		"TestID",
		"owasp.org",
	},
}

func TestNewGraph(t *testing.T) {
	g := NewGraph(NewCayleyGraphMemory())
	defer g.Close()

	t.Run("Testing NewGraph...", func(t *testing.T) {
		if g == nil {
			t.Errorf("Database is nil")
		}
	})

	t.Run("Testing db.String...", func(t *testing.T) {
		get := g.String()
		expected := g.db.String()

		if get != expected {
			t.Errorf("Error running String().\nGot %v\nWanted: %v", get, expected)
		}
	})
}
