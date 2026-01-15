package models

import (
	"fmt"
	"time"

	"github.com/example/orc/internal/db"
)

// Message represents an inter-agent message
type Message struct {
	ID        string
	Sender    string
	Recipient string
	Subject   string
	Body      string
	Timestamp time.Time
	Read      bool
	MissionID string
}

// CreateMessage creates a new message
func CreateMessage(sender, recipient, subject, body, missionID string) (*Message, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	// Verify mission exists
	var exists int
	err = database.QueryRow("SELECT COUNT(*) FROM missions WHERE id = ?", missionID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("mission %s not found", missionID)
	}

	// Generate message ID scoped to mission
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM messages WHERE mission_id = ?", missionID).Scan(&count)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("MSG-%s-%03d", missionID, count+1)

	// Insert message
	_, err = database.Exec(
		"INSERT INTO messages (id, sender, recipient, subject, body, mission_id, read) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, sender, recipient, subject, body, missionID, 0,
	)
	if err != nil {
		return nil, err
	}

	return GetMessage(id)
}

// GetMessage retrieves a message by ID
func GetMessage(id string) (*Message, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	message := &Message{}
	var readInt int
	err = database.QueryRow(
		"SELECT id, sender, recipient, subject, body, timestamp, read, mission_id FROM messages WHERE id = ?",
		id,
	).Scan(&message.ID, &message.Sender, &message.Recipient, &message.Subject, &message.Body, &message.Timestamp, &readInt, &message.MissionID)

	if err != nil {
		return nil, err
	}

	message.Read = readInt == 1

	return message, nil
}

// ListMessages retrieves messages for a recipient, optionally filtering to unread only
func ListMessages(recipient string, unreadOnly bool) ([]*Message, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := "SELECT id, sender, recipient, subject, body, timestamp, read, mission_id FROM messages WHERE recipient = ?"
	if unreadOnly {
		query += " AND read = 0"
	}
	query += " ORDER BY timestamp DESC"

	rows, err := database.Query(query, recipient)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		var readInt int
		err := rows.Scan(&message.ID, &message.Sender, &message.Recipient, &message.Subject, &message.Body, &message.Timestamp, &readInt, &message.MissionID)
		if err != nil {
			return nil, err
		}
		message.Read = readInt == 1
		messages = append(messages, message)
	}

	return messages, nil
}

// MarkRead marks a message as read
func MarkRead(id string) error {
	database, err := db.GetDB()
	if err != nil {
		return err
	}

	result, err := database.Exec("UPDATE messages SET read = 1 WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("message %s not found", id)
	}

	return nil
}

// GetConversation retrieves all messages between two agents, ordered by timestamp
func GetConversation(agent1, agent2 string) ([]*Message, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, sender, recipient, subject, body, timestamp, read, mission_id
		FROM messages
		WHERE (sender = ? AND recipient = ?) OR (sender = ? AND recipient = ?)
		ORDER BY timestamp ASC
	`

	rows, err := database.Query(query, agent1, agent2, agent2, agent1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		var readInt int
		err := rows.Scan(&message.ID, &message.Sender, &message.Recipient, &message.Subject, &message.Body, &message.Timestamp, &readInt, &message.MissionID)
		if err != nil {
			return nil, err
		}
		message.Read = readInt == 1
		messages = append(messages, message)
	}

	return messages, nil
}

// GetUnreadCount returns the count of unread messages for a recipient
func GetUnreadCount(recipient string) (int, error) {
	database, err := db.GetDB()
	if err != nil {
		return 0, err
	}

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM messages WHERE recipient = ? AND read = 0", recipient).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
