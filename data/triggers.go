package data

import (
	"database/sql"

	"github.com/PatronC2/Patron/types"
)

func ListTriggersByEvent(db *sql.DB, eventID int) ([]types.Trigger, error) {
	query := `SELECT "ID", "EventID", "ActionID" FROM "triggers" WHERE "EventID" = $1;`
	rows, err := db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var triggers []types.Trigger
	for rows.Next() {
		var t types.Trigger
		err := rows.Scan(&t.ID, &t.EventID, &t.ActionID)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, t)
	}

	return triggers, nil
}

func CreateTrigger(db *sql.DB, trigger types.Trigger) (int, error) {
	query := `INSERT INTO "triggers" ("EventID", "ActionID") VALUES ($1, $2) RETURNING "ID";`
	var triggerID int
	err := db.QueryRow(query, trigger.EventID, trigger.ActionID).Scan(&triggerID)
	if err != nil {
		return 0, err
	}
	return triggerID, nil
}

func DeleteTrigger(db *sql.DB, triggerID int) error {
	query := `DELETE FROM "triggers" WHERE "ID" = $1;`
	_, err := db.Exec(query, triggerID)
	return err
}
