// Package valueobjects contains value objects: small types defined by their
// value rather than an identity. SubscriptionStatus wraps a plain string in its
// own named type so the compiler can help us — a function expecting a
// SubscriptionStatus will not accept just any string by accident.
package valueobjects

// SubscriptionStatus is a named string type used as an "enum". Go has no
// built-in enum, so a block of typed constants is the idiomatic way to express
// a fixed set of allowed values.
type SubscriptionStatus string

const (
	StatusActive         SubscriptionStatus = "ACTIVE"
	StatusInactive       SubscriptionStatus = "INACTIVE"
	StatusCancelled      SubscriptionStatus = "CANCELLED"
	StatusPendingRenewal SubscriptionStatus = "PENDING_RENEWAL"
	StatusExpired        SubscriptionStatus = "EXPIRED"
)

// IsValidStatus reports whether a status is one of the known constants. It is
// used as a safety net when reading data from the outside (e.g. a status string
// loaded from the database) so an unexpected value can be detected and handled
// instead of silently flowing through the system. The switch lists every
// allowed case and the default catches anything else.
func IsValidStatus(status SubscriptionStatus) bool {
	switch status {
	case StatusActive, StatusInactive, StatusCancelled, StatusPendingRenewal, StatusExpired:
		return true
	default:
		return false
	}
}
