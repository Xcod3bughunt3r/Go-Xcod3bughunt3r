// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package netmap

import (
	"context"
	"net"
	"testing"
)

func TestNetblock(t *testing.T) {
	g := NewGraph(NewCayleyGraphMemory())
	for _, tt := range graphTest {
		t.Run("Testing UpsertNetblock...", func(t *testing.T) {
			got, err := g.UpsertNetblock(context.Background(), tt.CIDR, tt.Source, tt.EventID)
			if err != nil {
				t.Errorf("Error inserting netblock.\n%v\n", err)

			}

			get, _, err := net.ParseCIDR(got.(string))
			want, _, _ := net.ParseCIDR(tt.CIDR)

			if err != nil {
				t.Errorf("Error parsing node's cidr info from netblock.\n%v\n", got)
			}
			if !net.IP.Equal(get, want) {
				t.Errorf("Expected: %v\nGot: %v\n", want, get)
			}
		})

	}

}
