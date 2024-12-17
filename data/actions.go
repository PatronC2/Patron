package data

import (
	"database/sql"

	"github.com/PatronC2/Patron/types"
)

func ListActions(db *sql.DB) ([]types.Action, error) {
	query := `SELECT "action_id", "name", "description", "file" FROM "actions";`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []types.Action
	for rows.Next() {
		var a types.Action
		err := rows.Scan(&a.ActionID, &a.Name, &a.Description, &a.File)
		if err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}

	return actions, nil
}

func GetActionByID(db *sql.DB, eventID int) (types.Action, error) {
	query := `SELECT "action_id", "name", "description", "file" FROM "actions" WHERE "action_id" = $1;`
	var action types.Action
	err := db.QueryRow(query, eventID).Scan(
		&action.ActionID,
		&action.Name,
		&action.Description,
		&action.File,
	)
	if err != nil {
		return types.Action{}, err
	}
	return action, nil
}

func CreateAction(db *sql.DB, action types.Action) (int, error) {
	query := `INSERT INTO "actions" ("name", "description", "file") VALUES ($1, $2, $3) RETURNING "action_id";`
	var actionID int
	err := db.QueryRow(query, action.Name, action.Description, action.File).Scan(&actionID)
	if err != nil {
		return 0, err
	}
	return actionID, nil
}

func UpdateAction(db *sql.DB, action types.Action) error {
	query := `UPDATE "actions" SET "name" = $1, "description" = $2, "file" = $3 WHERE "action_id" = $4;`
	_, err := db.Exec(query, action.Name, action.Description, action.File, action.ActionID)
	return err
}

func DeleteAction(db *sql.DB, actionID int) error {
	query := `DELETE FROM "actions" WHERE "action_id" = $1;`
	_, err := db.Exec(query, actionID)
	return err
}
