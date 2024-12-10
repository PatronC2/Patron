package types

type Event struct {
	EventID     int     `db:"event_id"`
	Name        string  `db:"name"`
	Description string  `db:"description"`
	Script      []byte  `db:"script"`
	Schedule    string  `db:"schedule"`
	LastRun     *string `db:"last_run,omitempty"`
}

type Action struct {
	ActionID    int    `db:"action_id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	File        []byte `db:"file"`
}

type Trigger struct {
	ID       int `db:"id"`
	EventID  int `db:"event_id"`
	ActionID int `db:"action_id"`
}
