// Copyright 2022-07-21 Xcod3bughunt3r. All rights reserved.

package service

import "fmt"

// Service handles queued requests at an optional rate limit.
type Service interface {
	fmt.Stringer

	// Description returns a greeting message from the service.
	Description() string

	// Start requests that the service be started.
	Start() error

	// OnStart is called when the Start method requests the service be started.
	OnStart() error

	// Stop requests that the service be stopped.
	Stop() error

	// OnStop is called when the Stop method requests the service be stopped.
	OnStop() error

	// Done returns a channel that is closed when the service is stopped.
	Done() <-chan struct{}

	// Input returns a channel that the service receives requests on.
	Input() chan interface{}

	// Output returns a channel that the service send results on.
	Output() chan interface{}

	// SetRateLimit sets the number of calls to the OnRequest method each second.
	SetRateLimit(persec int)

	// CheckRateLimit blocks until the minimum wait duration since the last call.
	CheckRateLimit()
}
