package data

import (
	"database/sql"

	"github.com/PatronC2/Patron/types"
)

func ListEvents(db *sql.DB) ([]types.Event, error) {
	query := `SELECT "EventID", "Name", "Description", "Script", "Schedule", "LastRun" FROM "events";`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var e types.Event
		err := rows.Scan(&e.EventID, &e.Name, &e.Description, &e.Script, &e.Schedule, &e.LastRun)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func GetEventByID(db *sql.DB, eventID int) (types.Event, error) {
	query := `SELECT "EventID", "Name", "Description", "Script", "Schedule", "LastRun" FROM "events" WHERE "EventID" = $1;`
	var event types.Event
	err := db.QueryRow(query, eventID).Scan(
		&event.EventID,
		&event.Name,
		&event.Description,
		&event.Script,
		&event.Schedule,
		&event.LastRun,
	)
	if err != nil {
		return types.Event{}, err
	}
	return event, nil
}

func CreateEvent(db *sql.DB, event types.Event) (int, error) {
	query := `INSERT INTO "events" ("Name", "Description", "Script", "Schedule") VALUES ($1, $2, $3, $4) RETURNING "EventID";`
	var eventID int
	err := db.QueryRow(query, event.Name, event.Description, event.Script, event.Schedule).Scan(&eventID)
	if err != nil {
		return 0, err
	}
	return eventID, nil
}

func UpdateEvent(db *sql.DB, event types.Event) error {
	query := `UPDATE "events" SET "Name" = $1, "Description" = $2, "Script" = $3, "Schedule" = $4 WHERE "EventID" = $5;`
	_, err := db.Exec(query, event.Name, event.Description, event.Script, event.Schedule, event.EventID)
	return err
}

func DeleteEvent(db *sql.DB, eventID int) error {
	query := `DELETE FROM "events" WHERE "EventID" = $1;`
	_, err := db.Exec(query, eventID)
	return err
}
