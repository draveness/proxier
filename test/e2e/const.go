package e2e

import "time"

const (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 20
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)
