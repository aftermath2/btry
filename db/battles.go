package db

import (
	"database/sql"

	"github.com/aftermath2/BTRY/logger"

	"github.com/pkg/errors"
)

const (
	// BattleCreator is the role given to users that create a battle
	BattleCreator role = iota + 1
	// BattleChallenger is the role given to users that join a battle
	BattleChallenger
)

// Battle status
const (
	BattleCreated status = iota + 1
	BattleExpired
	BattleFinished
)

// BattlesStore contains the methods used to store and retrieve battles from the database.
type BattlesStore interface {
	AddPlayer(battleID uint64, player Player) error
	Create(battle Battle, player Player) error
	Expire(timestamp int64) ([]string, error)
	GetByID(id uint64) (Battle, error)
	GetPlayers(battleID uint64) ([]Player, error)
	List(offset, limit uint64, reverse bool) ([]Battle, error)
	Remove(id uint64) error
	Update(id uint64, status status, number uint64) error
}

type (
	role   uint8
	status uint8
)

// Player represents a user participating in a battle.
type Player struct {
	PublicKey string `json:"public_key,omitempty" db:"public_key"`
	Invoice   string `json:"invoice,omitempty"`
	Number    uint64 `json:"number,omitempty"`
	Role      role   `json:"role,omitempty"`
}

// Battle is a pvp game where users compete for guessing the closest number to a random one.
type Battle struct {
	Number    *uint64 `json:"number,omitempty"`
	CreatedAt uint64  `json:"created_at,omitempty" db:"created_at"`
	ID        uint64  `json:"id,omitempty"`
	Amount    uint64  `json:"amount,omitempty"`
	Status    status  `json:"status,omitempty"`
}

// battles ..
type battles struct {
	db     *sql.DB
	logger *logger.Logger
}

// newBattlesStore returns a new battles storage service.
func newBattlesStore(db *sql.DB, logger *logger.Logger) BattlesStore {
	return &battles{
		db:     db,
		logger: logger,
	}
}

// AddPlayer adds a player to a battle.
func (b *battles) AddPlayer(battleID uint64, player Player) error {
	if player.Role != BattleChallenger {
		return errors.New("invalid player, its role should be 'challenger'")
	}

	query := "INSERT INTO players (battle_id, role, public_key, invoice, number) VALUES (?,?,?,?,?)"
	stmt, err := b.db.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(battleID, player.Role, player.PublicKey, player.Invoice, player.Number); err != nil {
		return errors.Wrap(err, "updating battle")
	}

	return nil
}

// Create stores a new battle into the database.
func (b *battles) Create(battle Battle, player Player) error {
	if player.Role != BattleCreator {
		return errors.New("invalid player, its role should be 'creator'")
	}

	tx, err := b.db.Begin()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	battlesStmt, err := tx.Prepare("INSERT INTO battles (id, amount) VALUES (?, ?)")
	if err != nil {
		return errors.Wrap(err, "preparing insert statement")
	}
	defer battlesStmt.Close()

	if _, err := battlesStmt.Exec(battle.ID, battle.Amount); err != nil {
		return errors.Wrap(err, "adding battle")
	}

	query := "INSERT INTO players (battle_id, role, public_key, invoice, number) VALUES (?,?,?,?,?)"
	playersStmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "preparing insert statement")
	}
	defer playersStmt.Close()

	_, err = playersStmt.Exec(battle.ID, player.Role, player.PublicKey, player.Invoice, player.Number)
	if err != nil {
		return errors.Wrap(err, "adding battle")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

// Expire sets the battles lower than the timestamp as expired and deletes their players. It returns
// the invoices of the players so they can be cancelled.
func (b *battles) Expire(timestamp int64) ([]string, error) {
	tx, err := b.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	updateQuery := "UPDATE battles SET status=? WHERE status != ? AND created_at < ? RETURNING id"
	updateStmt, err := tx.Prepare(updateQuery)
	if err != nil {
		return nil, errors.Wrap(err, "preparing update statement")
	}
	defer updateStmt.Close()

	updateRows, err := updateStmt.Query(BattleExpired, BattleFinished, timestamp)
	if err != nil {
		return nil, err
	}
	defer updateRows.Close()

	ids, err := scanRows[uint64](updateRows)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	deleteQuery := "DELETE FROM players WHERE battle_id IN "
	deleteQuery += BulkInsertValues(1, len(ids))
	deleteQuery += " RETURNING invoice"

	deleteStmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		return nil, errors.Wrap(err, "preparing delete statement")
	}
	defer deleteStmt.Close()

	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	deleteRows, err := deleteStmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer deleteRows.Close()

	invoices, err := scanRows[string](deleteRows)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	return invoices, nil
}

// GetByID returns the battle with the ID provided.
func (b *battles) GetByID(id uint64) (Battle, error) {
	stmt, err := b.db.Prepare("SELECT id, amount, status, number, created_at FROM battles WHERE id=?")
	if err != nil {
		return Battle{}, errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	var battle Battle
	row := stmt.QueryRow(id)
	err = row.Scan(&battle.ID, &battle.Amount, &battle.Status, &battle.Number, &battle.CreatedAt)
	if err != nil {
		return Battle{}, errors.Wrap(err, "scanning battle")
	}

	return battle, nil
}

func (b *battles) GetPlayers(battleID uint64) ([]Player, error) {
	query := "SELECT role, public_key, invoice, number FROM players WHERE battle_id=?"
	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "preparing players statement")
	}
	defer stmt.Close()

	rows, err := stmt.Query(battleID)
	if err != nil {
		return nil, errors.Wrap(err, "querying players")
	}
	defer rows.Close()

	var (
		players []Player
		// Reuse object
		player Player
	)
	for rows.Next() {
		if err := rows.Scan(&player.Role, &player.PublicKey, &player.Invoice, &player.Number); err != nil {
			return nil, errors.Wrap(err, "scanning player")
		}

		players = append(players, player)
	}

	return players, nil
}

// List returns a list of battles stored in the database.
func (b *battles) List(offset, limit uint64, reverse bool) ([]Battle, error) {
	// Cap limit to avoid creating a slice with too big capacity
	if limit > 100 {
		limit = 100
	}

	query := "SELECT id, amount, status, number, created_at FROM battles"
	clauses := AddPagination(offset, limit, "id", reverse)
	query += clauses

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(offset, offset, limit)
	if err != nil {
		return nil, errors.Wrap(err, "listing battles")
	}
	defer rows.Close()

	battles := make([]Battle, 0, limit)
	// Reuse object
	var battle Battle
	for rows.Next() {
		err := rows.Scan(&battle.ID, &battle.Amount, &battle.Status, &battle.Number, &battle.CreatedAt)
		if err != nil {
			return nil, err
		}

		battles = append(battles, battle)
	}

	return battles, nil
}

// Remove deletes a battle from the database.
func (b *battles) Remove(id uint64) error {
	stmt, err := b.db.Prepare("DELETE FROM battles WHERE id=?")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		return errors.Wrap(err, "deleting battle")
	}

	return nil
}

// Update updates a battle's status and number.
func (b *battles) Update(id uint64, status status, number uint64) error {
	stmt, err := b.db.Prepare("UPDATE battles SET status=?, number=? WHERE id=?")
	if err != nil {
		return errors.Wrap(err, "preparing statement")
	}
	defer stmt.Close()

	if _, err := stmt.Exec(status, number, id); err != nil {
		return errors.Wrap(err, "updating battle")
	}

	return nil
}
