package db_test

import (
	"database/sql"
	"testing"

	"github.com/aftermath2/BTRY/db"
	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/suite"
)

var lightningAddress = "test@domain.com"

type LightningSuite struct {
	suite.Suite

	db database.LightningStore
}

func TestLightningSuite(t *testing.T) {
	suite.Run(t, &LightningSuite{})
}

func (l *LightningSuite) SetupTest() {
	db := setupDB(l.T(), func(db *sql.DB) {
		query := `DELETE FROM lightning;
		INSERT INTO lightning (public_key, address) VALUES (?, ?);`
		_, err := db.Exec(query, testWinner.PublicKey, lightningAddress)
		l.NoError(err)
	})
	l.db = db.Lightning
}

func (l *LightningSuite) TestGetAddress() {
	address, err := l.db.GetAddress(testWinner.PublicKey)
	l.NoError(err)

	l.Equal(lightningAddress, address)
}

func (l *LightningSuite) TestGetAddressNoAddress() {
	_, err := l.db.GetAddress("random")
	l.ErrorIs(err, db.ErrNoAddress)
}

func (l *LightningSuite) TestSetAddress() {
	publicKey := "pubKey"
	address := "test2@domain.com"
	err := l.db.SetAddress(publicKey, address)
	l.NoError(err)

	gotAddress, err := l.db.GetAddress(publicKey)
	l.NoError(err)

	l.Equal(address, gotAddress)
}
