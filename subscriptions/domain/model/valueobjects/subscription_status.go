package valueobjects

type SubscriptionStatus string

const (
	StatusActive         SubscriptionStatus = "ACTIVE"
	StatusInactive       SubscriptionStatus = "INACTIVE"
	StatusCancelled      SubscriptionStatus = "CANCELLED"
	StatusPendingRenewal SubscriptionStatus = "PENDING_RENEWAL"
	StatusExpired        SubscriptionStatus = "EXPIRED"
)

func IsValidStatus(status SubscriptionStatus) bool {
	switch status {
	case StatusActive, StatusInactive, StatusCancelled, StatusPendingRenewal, StatusExpired:
		return true
	default:
		return false
	}
}
