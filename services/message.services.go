package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/emarifer/go-echo-templ-htmx/db"
)

func NewMessageServices(m Message, mStore db.Store) *MessageServices {
	return &MessageServices{
		Message:      m,
		MessageStore: mStore,
	}
}

type Message struct {
	ID                  int         `json:"id"`
	NotificationGroupID int         `json:"group_id"`
	Topic               string      `json:"topic"`
	Description         string      `json:"description"`
	SendTime            time.Time   `json:"send_time"`
	ReadStatus          map[int]int `json:"read_status"`
	CreatedBy           int         `json:"created_by"`
	CreatedAt           time.Time   `json:"created_at,omitempty"`
}

// { All userIds of recipients in the notificaiton group: unseen }

type MessageServices struct {
	Message      Message
	MessageStore db.Store
}

func (ms *MessageServices) CreateMessage(m Message) (Message, error) {
	query := `INSERT INTO messages (notification_group_id, topic, description, send_time, created_by)
		VALUES(?, ?, ?, ?, ?) RETURNING *`

	stmt, err := ms.MessageStore.Db.Prepare(query)
	if err != nil {
		return Message{}, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(
		m.NotificationGroupID,
		m.Topic,
		m.Description,
		m.SendTime,
		m.CreatedBy,
	).Scan(
		&ms.Message.ID,
		&ms.Message.NotificationGroupID,
		&ms.Message.Topic,
		&ms.Message.Description,
		&ms.Message.SendTime,
		&ms.Message.ReadStatus,
		&ms.Message.CreatedBy,
		&ms.Message.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}

	return ms.Message, nil
}

func (ms *MessageServices) GetAllMessagesByNotificationGroup(notificationGroupID int) ([]Message, error) {
	query := fmt.Sprintf("SELECT id, topic, send_time, read_status, description, created_by FROM messages WHERE notification_group_id = %d ORDER BY created_at DESC", notificationGroupID)

	rows, err := ms.MessageStore.Db.Query(query)
	if err != nil {
		return []Message{}, err
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Topic, &m.SendTime, &m.ReadStatus, &m.Description, &m.CreatedBy); err != nil {
			continue // handle error or break as necessary
		}
		messages = append(messages, m)
	}

	return messages, nil
}

func (ms *MessageServices) GetMessageById(id int) (Message, error) {
	query := `SELECT * FROM messages WHERE id=?`

	stmt, err := ms.MessageStore.Db.Prepare(query)
	if err != nil {
		return Message{}, err
	}

	defer stmt.Close()

	var m Message
	err = stmt.QueryRow(id).Scan(
		&m.ID,
		&m.NotificationGroupID,
		&m.Topic,
		&m.Description,
		&m.SendTime,
		&m.ReadStatus,
		&m.CreatedBy,
		&m.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}

	return m, nil
}

func (ms *MessageServices) UpdateMessage(m Message) (Message, error) {
	query := `UPDATE messages SET topic = ?, description = ?, send_time = ?
		WHERE id = ? RETURNING *`

	stmt, err := ms.MessageStore.Db.Prepare(query)
	if err != nil {
		return Message{}, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(
		m.Topic,
		m.Description,
		m.SendTime,
		m.ID,
	).Scan(
		&m.ID,
		&m.NotificationGroupID,
		&m.Topic,
		&m.Description,
		&m.SendTime,
		&m.ReadStatus,
		&m.CreatedBy,
		&m.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}

	return m, nil
}

func (ms *MessageServices) DeleteMessage(id int) error {
	query := `DELETE FROM messages WHERE id = ?`

	stmt, err := ms.MessageStore.Db.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return err
	}

	if i, err := result.RowsAffected(); err != nil || i != 1 {
		return errors.New("an affected row was expected")
	}

	return nil
}
