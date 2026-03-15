package scraper

import "errors"

// Sentinel errors as required by the modulemaker contract.
var (
	ErrInvalidURL      = errors.New("facebook: invalid URL")
	ErrNotFound        = errors.New("facebook: resource not found")
	ErrPrivateResource = errors.New("facebook: resource is private")
	ErrRateLimited     = errors.New("facebook: rate limited")
	ErrBlocked         = errors.New("facebook: blocked by upstream")
	ErrUpstreamChanged = errors.New("facebook: upstream response format changed")
	ErrContextCanceled = errors.New("facebook: context canceled")
)
