package counter

const (
	incrQuery = `INSERT INTO counters (key, bucket, count) VALUES (?, ?, 1)
		ON CONFLICT (key, bucket) DO UPDATE SET count = counters.count + 1
		RETURNING count`

	deleteBeforeQuery = "DELETE FROM counters WHERE bucket < ?"
)
