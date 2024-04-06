package db_test

import (
	"database/sql"
	"testing"

	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/suite"
)

const (
	notificationPublicKey       = "7d959d6d552c7d38b3ecafb72805fa03a6dee6b7f0c5f63f57a371736cb004b1"
	notificationChatID    int64 = 505
)

type NotificationsSuite struct {
	suite.Suite

	db database.NotificationsStore
}

func TestNotificationsSuite(t *testing.T) {
	suite.Run(t, &NotificationsSuite{})
}

func (n *NotificationsSuite) SetupTest() {
	db := setupDB(n.T(), func(db *sql.DB) {
		query := `DELETE FROM notifications;
		INSERT INTO notifications (public_key, chat_id, service) VALUES (?, ?, ?);`
		_, err := db.Exec(query, notificationPublicKey, notificationChatID, "telegram")
		n.NoError(err)
	})
	n.db = db.Notifications
}

func (n *NotificationsSuite) TestAdd() {
	publicKey := "876baf90c3d2d26c04ba1d208c29605b2c6fd13fbb3f6b46cf7f10ece3dac69d"
	chatID := int64(132009)
	err := n.db.Add(publicKey, chatID)
	n.NoError(err)

	_, err = n.db.GetChatID(publicKey)
	n.NoError(err)
}

func (n *NotificationsSuite) TestGetChatID() {
	gotChatID, err := n.db.GetChatID(notificationPublicKey)
	n.NoError(err)

	n.Equal(notificationChatID, gotChatID)
}
