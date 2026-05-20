package dto

type NotificationEvent struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}
