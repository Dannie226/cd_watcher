package queries

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetMaxVersionID(conn *pgx.Conn) (pgtype.Int8, error) {
	row := conn.QueryRow(getTimeoutContext(), "select max(id) from versions;")

	var id pgtype.Int8
	err := row.Scan(&id)

	return id, err
}

func GetNextVersionID(conn *pgx.Conn) (int64, error) {
	id, err := GetMaxVersionID(conn)

	if err != nil {
		return 0, err
	}

	if id.Valid {
		return id.Int64 + 1, nil
	}

	return 0, nil
}
