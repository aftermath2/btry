package db_test

import (
	"database/sql"
	"testing"
	"time"

	database "github.com/aftermath2/BTRY/db"

	"github.com/stretchr/testify/assert"
)

var (
	testBattle = database.Battle{
		ID:     324,
		Amount: 50000,
		Status: database.BattleCreated,
	}

	testBattle2 = database.Battle{
		ID:     5783,
		Amount: 99999,
		Status: database.BattleCreated,
	}

	testBattle3 = database.Battle{
		ID:     9463,
		Amount: 12454,
		Status: database.BattleCreated,
	}

	testPlayer = database.Player{
		PublicKey: "7d959d6d552c7d38b3ecafb72805fa03a6dee6b7f0c5f63f57a371736cb004b1",
		Invoice:   "lnbrc",
		Number:    506,
		Role:      database.BattleCreator,
	}

	testPlayer2 = database.Player{
		PublicKey: "876baf90c3d2d26c04ba1d208c29605b2c6fd13fbb3f6b46cf7f10ece3dac69d",
		Invoice:   "lnbrc",
		Number:    905,
		Role:      database.BattleCreator,
	}

	testPlayer3 = database.Player{
		PublicKey: "876baf90c3d2d26c04ba1d208c29605b2c6fd13fbb3f6b46cf7f10ece3dac69d",
		Invoice:   "lnbrc",
		Number:    746,
		Role:      database.BattleChallenger,
	}
)

func TestAddPlayer(t *testing.T) {
	db := setupBattles(t)

	err := db.Battles.AddPlayer(testBattle.ID, testPlayer3)
	assert.NoError(t, err)

	players, err := db.Battles.GetPlayers(testBattle.ID)
	assert.NoError(t, err)

	assert.Len(t, players, 2)
	assert.Equal(t, testPlayer, players[0])
	assert.Equal(t, testPlayer3, players[1])
}

func TestAddPlayerErrors(t *testing.T) {
	db := setupBattles(t)

	t.Run("Invalid player role", func(t *testing.T) {
		err := db.Battles.AddPlayer(testBattle.ID, testPlayer2)
		assert.Error(t, err)
	})
}

func TestCreateBattle(t *testing.T) {
	db := setupBattles(t)

	err := db.Battles.Create(testBattle3, testPlayer2)
	assert.NoError(t, err)

	battle, err := db.Battles.GetByID(testBattle3.ID)
	assert.NoError(t, err)

	assertEqualBattle(t, testBattle3, battle)

	players, err := db.Battles.GetPlayers(testBattle3.ID)
	assert.NoError(t, err)

	assert.Len(t, players, 1)
	assert.Equal(t, testPlayer2, players[0])
}

func TestGetByID(t *testing.T) {
	db := setupBattles(t)

	battle, err := db.Battles.GetByID(testBattle.ID)
	assert.NoError(t, err)

	assertEqualBattle(t, testBattle, battle)
}

func TestListBattles(t *testing.T) {
	db := setupBattles(t)

	cases := []struct {
		desc     string
		expected []database.Battle
		offset   uint64
		limit    uint64
	}{
		{
			desc:     "Standard",
			expected: []database.Battle{testBattle, testBattle2},
		},
		{
			desc:     "Offset",
			offset:   testBattle.ID,
			expected: []database.Battle{testBattle2},
		},
		{
			desc:     "Limit",
			limit:    1,
			expected: []database.Battle{testBattle},
		},
		{
			desc:     "Offset & limit",
			limit:    1,
			offset:   testBattle.ID,
			expected: []database.Battle{testBattle2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			battles, err := db.Battles.List(tc.offset, tc.limit, false)
			assert.NoError(t, err)

			for i, battle := range battles {
				assertEqualBattle(t, tc.expected[i], battle)
			}
		})
	}
}

func TestExpire(t *testing.T) {
	db := setupBattles(t)

	invoices, err := db.Battles.Expire(time.Now().Add(time.Hour).Unix())
	assert.NoError(t, err)

	assert.Len(t, invoices, 1)
	assert.Equal(t, testPlayer.Invoice, invoices[0])

	players, err := db.Battles.GetPlayers(testBattle.ID)
	assert.NoError(t, err)

	assert.Len(t, players, 0)
}

func TestRemoveBattle(t *testing.T) {
	db := setupBattles(t)

	err := db.Battles.Remove(testBattle.ID)
	assert.NoError(t, err)

	_, err = db.Battles.GetByID(testBattle.ID)
	assert.Error(t, err)
}

func TestUpdateBattle(t *testing.T) {
	db := setupBattles(t)

	status := database.BattleFinished
	number := uint64(100)
	err := db.Battles.Update(testBattle.ID, status, number)
	assert.NoError(t, err)

	battle, err := db.Battles.GetByID(testBattle.ID)
	assert.NoError(t, err)

	assert.Equal(t, status, battle.Status)
	assert.Equal(t, number, *battle.Number)
}

func assertEqualBattle(t *testing.T, expected, actual database.Battle) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Amount, actual.Amount)
	assert.Equal(t, expected.Status, actual.Status)
	assert.Equal(t, expected.Number, actual.Number)
}

func setupBattles(t *testing.T) *database.DB {
	t.Helper()

	return setupDB(t, func(db *sql.DB) {
		battlesQuery := `DELETE FROM battles;
		INSERT INTO battles (id, amount) VALUES (?,?), (?,?);`
		_, err := db.Exec(battlesQuery,
			testBattle.ID, testBattle.Amount,
			testBattle2.ID, testBattle2.Amount,
		)
		assert.NoError(t, err)

		playersQuery := `DELETE FROM players;
		INSERT INTO players (battle_id, public_key, role, invoice, number) VALUES (?,?,?,?,?);`
		_, err = db.Exec(playersQuery, testBattle.ID, testPlayer.PublicKey, testPlayer.Role,
			testPlayer.Invoice, testPlayer.Number)
		assert.NoError(t, err)
	})
}
