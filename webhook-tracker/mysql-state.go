package webhook_tracker

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type MySqlState struct {
	dbConn *sql.DB
}

func NewMySqlState(db_url string) *MySqlState {
	db, err := sql.Open("mysql", db_url)

	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	log.Println("Connected to MySQL")

	return &MySqlState{
		dbConn: db,
	}
}

func (m *MySqlState) IncrementCallCount(webhook string, queryID string) int64 {
	// Use a mysql upsert to set the counter to 1 if it doesn't exist, otherwise increment it
	// This should be done in a transaction
	tx, err := m.dbConn.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %s", err)
		return 0
	}
	defer func(tx *sql.Tx) {
		_ = tx.Commit()
	}(tx)
	// First try to update the counter
	stmt, err := tx.Prepare("INSERT INTO `webhook_stats` (`webhook`, `query_id`, `invoke_count`) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE `invoke_count` = `invoke_count` + 1")
	if err != nil {
		log.Printf("Error preparing statement: %s", err)
		return 0
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)
	_, err = stmt.Exec(webhook, queryID)

	if err != nil {
		log.Printf("Error executing statement: %s", err)
		return 0
	}

	// Now get the value
	var value int64
	err = tx.QueryRow("SELECT `invoke_count` FROM `webhook_stats` WHERE `webhook` = ? AND `query_id` = ?", webhook, queryID).Scan(&value)
	if err != nil {
		log.Printf("Error getting value: %s", err)
		return 0
	}
	return value
}

func (m *MySqlState) HasBeenCalled(webhook string, queryID string) bool {
	// If the invoke count is greater than 0, return true, otherwise false
	var value int64
	err := m.dbConn.QueryRow("SELECT `invoke_count` FROM `webhook_stats` WHERE `webhook` = ? AND `query_id` = ?", webhook, queryID).Scan(&value)
	if err != nil {
		log.Printf("Error getting value: %s", err)
		return false
	}
	return value > 0

}
