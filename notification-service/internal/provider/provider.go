package provider

// NotificationProvider is the adapter interface for sending notifications.
type NotificationProvider interface {
	Send(to, subject, body string) error
}
