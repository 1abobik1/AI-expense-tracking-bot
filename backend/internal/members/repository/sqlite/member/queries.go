package member

const (
	columns = "user_id, username, first_name, name, status, created_at"

	// upsertQuery пишет, только если имя изменилось: вызывается на каждое сообщение, а записи D1 лимитированы.
	upsertQuery = `INSERT INTO members (user_id, username, first_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE
		SET username = excluded.username, first_name = excluded.first_name, updated_at = excluded.updated_at
		WHERE members.username <> excluded.username OR members.first_name <> excluded.first_name`

	updateQuery = "UPDATE members SET name = ?, status = ?, updated_at = ? WHERE user_id = ?"

	getQuery = "SELECT " + columns + " FROM members WHERE user_id = ?"

	listQuery = "SELECT " + columns + " FROM members ORDER BY created_at, user_id"
)
