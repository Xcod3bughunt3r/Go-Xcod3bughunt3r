// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package resolve

import (
	"testing"
)

func TestClientSubnetCheck(t *testing.T) {
	good := []string{
		"8.8.8.8:53",     // Google
		"1.1.1.1:53",     // Cloudflare
		"209.244.0.3:53", // Level3
	}
	bad := []string{
		"208.76.50.50:53", // SmartViper
	}

	for _, r := range good {
		if err := ClientSubnetCheck(r); err != nil {
			t.Errorf("%v", err)
		}
	}
	for _, r := range bad {
		if err := ClientSubnetCheck(r); err == nil {
			t.Errorf("%s should have failed the test", r)
		}
	}
}
