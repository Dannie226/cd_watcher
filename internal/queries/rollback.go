package queries

import (
	"github.com/jackc/pgx/v5"
)

func GetLastVersions(conn *pgx.Conn, n int) ([]VersionInfo, error) {
	rows, err := conn.Query(
		getTimeoutContext(),
		"select id, foldername from versions order by id desc limit $1",
		n,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	i := make([]VersionInfo, 0, n)

	for rows.Next() {
		inf := VersionInfo{}

		if err := rows.Scan(
			&inf.ID,
			&inf.FolderName,
		); err != nil {
			return nil, err
		}

		i = append(i, inf)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return i, nil
}

func RemoveVersions(conn *pgx.Conn, n int) error {
	_, err := conn.Exec(
		getTimeoutContext(),
		"delete from versions where id>(select max(id)-$1 from versions)",
		n,
	)

	return err
}
