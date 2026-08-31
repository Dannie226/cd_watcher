package queries

import (
	"github.com/jackc/pgx/v5"
)

func InsertNewVersion(conn *pgx.Conn, params VersionInfo) error {
	_, err := conn.Exec(
		getTimeoutContext(),
		"insert into versions(id, foldername) values ($1, $2)",
		params.ID,
		params.FolderName,
	)

	return err
}
