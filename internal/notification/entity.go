package notification

import "time"

type Notification struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Body      string
	IsRead    bool
	CreatedAt time.Time
}

func (n *Notification) IsUnread() bool {
	return !n.IsRead
}
