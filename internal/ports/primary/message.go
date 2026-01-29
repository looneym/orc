package primary

import "context"

// MessageService defines the primary port for message operations.
type MessageService interface {
	// CreateMessage creates a new message.
	CreateMessage(ctx context.Context, req CreateMessageRequest) (*CreateMessageResponse, error)

	// GetMessage retrieves a message by ID.
	GetMessage(ctx context.Context, messageID string) (*Message, error)

	// ListMessages lists messages for a recipient.
	ListMessages(ctx context.Context, recipient string, unreadOnly bool) ([]*Message, error)

	// MarkRead marks a message as read.
	MarkRead(ctx context.Context, messageID string) error

	// GetConversation retrieves all messages between two actors.
	GetConversation(ctx context.Context, actor1, actor2 string) ([]*Message, error)

	// GetUnreadCount returns the count of unread messages for a recipient.
	GetUnreadCount(ctx context.Context, recipient string) (int, error)
}

// CreateMessageRequest contains parameters for creating a message.
// Sender and Recipient are actor IDs (BENCH-xxx, GATE-xxx, WATCH-xxx).
type CreateMessageRequest struct {
	Sender    string
	Recipient string
	Subject   string
	Body      string
}

// CreateMessageResponse contains the result of creating a message.
type CreateMessageResponse struct {
	MessageID string
	Message   *Message
}

// Message represents a message entity at the port boundary.
// Sender and Recipient are actor IDs (BENCH-xxx, GATE-xxx, WATCH-xxx).
type Message struct {
	ID        string
	Sender    string
	Recipient string
	Subject   string
	Body      string
	Timestamp string
	Read      bool
}
