// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package resolve

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

const thresholdCheckInterval time.Duration = 3 * time.Second

type ThresholdOptions struct {
	ThresholdValue         uint64
	CumulativeAccumulation bool // instead of continuous
	CountTimeouts          bool
	CountFormatErrors      bool
	CountServerFailures    bool
	CountNotImplemented    bool
	CountQueryRefusals     bool
}

type stats struct {
	sync.Mutex
	LastSuccess         uint64
	CountTimeouts       bool
	Timeouts            uint64
	CountFormatErrors   bool
	FormatErrors        uint64
	CountServerFailures bool
	ServerFailures      uint64
	CountNotImplemented bool
	NotImplemented      uint64
	CountQueryRefusals  bool
	QueryRefusals       uint64
}

// SetThresholdOptions updates the settings used for discontinuing use of a resolver due to poor performance.
func (r *Resolvers) SetThresholdOptions(opt *ThresholdOptions) {
	r.Lock()
	defer r.Unlock()

	r.options = opt
	r.updateThresholdOptions()
}

func (r *Resolvers) updateThresholdOptions() {
	for _, res := range r.list {
		select {
		case <-res.done:
		default:
			res.stats.Lock()
			res.stats.CountTimeouts = r.options.CountTimeouts
			res.stats.CountFormatErrors = r.options.CountFormatErrors
			res.stats.CountServerFailures = r.options.CountServerFailures
			res.stats.CountNotImplemented = r.options.CountNotImplemented
			res.stats.CountQueryRefusals = r.options.CountQueryRefusals
			res.stats.Unlock()
		}
	}
}

func (r *Resolvers) thresholdChecks() {
	t := time.NewTicker(thresholdCheckInterval)
	defer t.Stop()

	for {
		select {
		case <-r.done:
			return
		case <-t.C:
			r.shutdownIfThresholdViolated()
		}
	}
}

func (r *Resolvers) shutdownIfThresholdViolated() {
	r.Lock()
	list := r.list
	opts := *r.options
	r.Unlock()

	tv := opts.ThresholdValue
	if tv == 0 {
		return
	}

	for idx, res := range list {
		var stop bool

		if opts.CumulativeAccumulation && res.cumulativeThresholdReached(tv) {
			stop = true
		} else if res.continuousThresholdReached(tv) {
			stop = true
		}
		if stop {
			r.stopResolver(idx)
		}
	}
}

func (r *resolver) continuousThresholdReached(tv uint64) bool {
	r.stats.Lock()
	defer r.stats.Unlock()

	return r.stats.LastSuccess >= tv
}

func (r *resolver) cumulativeThresholdReached(tv uint64) bool {
	r.stats.Lock()
	defer r.stats.Unlock()

	var total uint64
	if r.stats.CountTimeouts {
		total += r.stats.Timeouts
	}
	if r.stats.CountFormatErrors {
		total += r.stats.FormatErrors
	}
	if r.stats.CountServerFailures {
		total += r.stats.ServerFailures
	}
	if r.stats.CountNotImplemented {
		total += r.stats.NotImplemented
	}
	if r.stats.CountQueryRefusals {
		total += r.stats.QueryRefusals
	}
	return total >= tv
}

func (r *resolver) collectStats(resp *dns.Msg) {
	if resp == nil {
		return
	}

	r.stats.Lock()
	defer r.stats.Unlock()

	switch resp.Rcode {
	case RcodeNoResponse:
		r.stats.Timeouts++
		if r.stats.CountTimeouts {
			r.stats.LastSuccess++
		}
	case dns.RcodeFormatError:
		r.stats.FormatErrors++
		if r.stats.CountFormatErrors {
			r.stats.LastSuccess++
		}
	case dns.RcodeServerFailure:
		r.stats.ServerFailures++
		if r.stats.CountServerFailures {
			r.stats.LastSuccess++
		}
	case dns.RcodeNotImplemented:
		r.stats.NotImplemented++
		if r.stats.CountNotImplemented {
			r.stats.LastSuccess++
		}
	case dns.RcodeRefused:
		r.stats.QueryRefusals++
		if r.stats.CountQueryRefusals {
			r.stats.LastSuccess++
		}
	default:
		r.stats.LastSuccess = 0
	}
}
