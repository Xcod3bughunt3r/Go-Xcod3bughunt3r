// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package resolve

import (
	"strings"
	"testing"
	"time"

	"github.com/caffix/stringset"
	"github.com/miekg/dns"
)

func TestXchgAddRemove(t *testing.T) {
	name := "caffix.net"
	xchg := newXchgMgr(DefaultTimeout)
	msg := QueryMsg(name, dns.TypeA)
	req := &request{
		ID:    msg.Id,
		Name:  name,
		Qtype: dns.TypeA,
		Msg:   msg,
	}
	if err := xchg.add(req); err != nil {
		t.Errorf("Failed to add the request")
	}
	if err := xchg.add(req); err == nil {
		t.Errorf("Failed to detect the same request added twice")
	}

	ret := xchg.remove(msg.Id, msg.Question[0].Name)
	if ret == nil || ret.Msg == nil || name != strings.ToLower(RemoveLastDot(ret.Msg.Question[0].Name)) {
		t.Errorf("Did not find and remove the message from the data structure")
	}
	ret = xchg.remove(msg.Id, msg.Question[0].Name)
	if ret != nil {
		t.Errorf("Did not return nil when attempting to remove an element for the second time")
	}
	if err := xchg.add(req); err != nil {
		t.Errorf("Failed to add the request after being removed")
	}
}

func TestXchgUpdateTimestamp(t *testing.T) {
	name := "caffix.net"
	xchg := newXchgMgr(DefaultTimeout)
	msg := QueryMsg(name, dns.TypeA)

	req := &request{
		ID:    msg.Id,
		Name:  name,
		Qtype: dns.TypeA,
		Msg:   msg,
	}

	if !req.Timestamp.IsZero() {
		t.Errorf("Expected the new request to have a zero value timestamp")
	}
	if err := xchg.add(req); err != nil {
		t.Errorf("Failed to add the request")
	}
	xchg.updateTimestamp(msg.Id, name)
	// For complete coverage
	xchg.updateTimestamp(msg.Id, "Bad Name")

	req = xchg.remove(msg.Id, msg.Question[0].Name)
	if req == nil || req.Timestamp.IsZero() {
		t.Errorf("Expected the updated request to not have a zero value timestamp")
	}
}

func TestXchgRemoveExpired(t *testing.T) {
	xchg := newXchgMgr(time.Second)
	names := []string{"caffix.net", "www.caffix.net", "blog.caffix.net"}

	for _, name := range names {
		msg := QueryMsg(name, dns.TypeA)
		if err := xchg.add(&request{
			ID:        msg.Id,
			Name:      name,
			Qtype:     dns.TypeA,
			Msg:       msg,
			Timestamp: time.Now(),
		}); err != nil {
			t.Errorf("Failed to add the request")
		}
	}
	// Add one request that should not be removed with the others
	name := "vpn.caffix.net"
	msg := QueryMsg(name, dns.TypeA)
	if err := xchg.add(&request{
		ID:        msg.Id,
		Name:      name,
		Qtype:     dns.TypeA,
		Msg:       msg,
		Timestamp: time.Now().Add(3 * time.Second),
	}); err != nil {
		t.Errorf("Failed to add the request")
	}
	if len(xchg.removeExpired()) > 0 {
		t.Errorf("The removeExpired method returned requests too early")
	}

	time.Sleep(1500 * time.Millisecond)
	set := stringset.New(names...)
	defer set.Close()

	for _, req := range xchg.removeExpired() {
		set.Remove(req.Name)
	}
	if set.Len() > 0 {
		t.Errorf("Not all expected requests were returned by removeExpired")
	}
}

func TestXchgRemoveAll(t *testing.T) {
	xchg := newXchgMgr(time.Second)
	names := []string{"caffix.net", "www.caffix.net", "blog.caffix.net"}

	for _, name := range names {
		msg := QueryMsg(name, dns.TypeA)
		if err := xchg.add(&request{
			ID:    msg.Id,
			Name:  name,
			Qtype: dns.TypeA,
			Msg:   msg,
		}); err != nil {
			t.Errorf("Failed to add the request")
		}
	}

	set := stringset.New(names...)
	defer set.Close()

	for _, req := range xchg.removeAll() {
		set.Remove(req.Name)
	}
	if set.Len() > 0 {
		t.Errorf("Not all expected requests were returned by removeAll")
	}
}
