package expense

const (
	columns = "id, user_id, amount, category, spent_at, comment, created_at, updated_at"

	insertQuery = `INSERT INTO expenses (user_id, amount, category, spent_at, comment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id`

	selectQuery = "SELECT " + columns + " FROM expenses"

	summaryQuery = "SELECT category, COUNT(*), SUM(amount) FROM expenses"

	updateQuery = "UPDATE expenses SET amount = ?, category = ?, spent_at = ?, comment = ?, updated_at = ?"

	deleteQuery = "DELETE FROM expenses"

	orderPage = " ORDER BY spent_at DESC, id DESC LIMIT ? OFFSET ?"
)
