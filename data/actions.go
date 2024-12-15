package data

import (
	"database/sql"

	"github.com/PatronC2/Patron/types"
)

func ListActions(db *sql.DB) ([]types.Action, error) {
	query := `SELECT "ActionID", "Name", "Description", "File" FROM "actions";`
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

func CreateAction(db *sql.DB, action types.Action) (int, error) {
	query := `INSERT INTO "actions" ("Name", "Description", "File") VALUES ($1, $2, $3) RETURNING "ActionID";`
	var actionID int
	err := db.QueryRow(query, action.Name, action.Description, action.File).Scan(&actionID)
	if err != nil {
		return 0, err
	}
	return actionID, nil
}

func UpdateAction(db *sql.DB, action types.Action) error {
	query := `UPDATE "actions" SET "Name" = $1, "Description" = $2, "File" = $3 WHERE "ActionID" = $4;`
	_, err := db.Exec(query, action.Name, action.Description, action.File, action.ActionID)
	return err
}

func DeleteAction(db *sql.DB, actionID int) error {
	query := `DELETE FROM "actions" WHERE "ActionID" = $1;`
	_, err := db.Exec(query, actionID)
	return err
}
