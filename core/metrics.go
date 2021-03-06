package core

// MetricsFactory is a factory for configuring the metrics for the environment.
type MetricsFactory interface {
	Configure(*Environment) error
}
