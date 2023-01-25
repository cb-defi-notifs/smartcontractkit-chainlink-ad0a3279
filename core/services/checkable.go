package services

// Checkable should be implemented by any type requiring health checks.
// From the k8s docs:
// > ready means it’s initialized and healthy means that it can accept traffic in kubernetes
// See: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/
type Checkable interface {
	// Ready should return nil if ready, or an error message otherwise.
	Ready() error
	// Healthy should return nil if healthy, or an error message otherwise.
	Healthy() error
}
