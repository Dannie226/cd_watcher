package queries

import (
	"errors"
	"fmt"

	"github.com/Dannie226/cd_watcher/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmailIds struct {
	Main  pgtype.Text
	Chain pgtype.Text
	Event int32
}

var ErrUnknownEvent error = errors.New("Unknown email event type")

func GetEmailChainIDs(conn *pgx.Conn, event config.EmailEvent) (EmailIds, error) {
	var ids EmailIds

	switch {
	case event.RecievesEmail(config.DeployEvent):
		ids.Event = 1

	case event.RecievesEmail(config.RollbackEvent):
		ids.Event = 2
	}

	if ids.Event == 0 {
		return EmailIds{}, ErrUnknownEvent
	}

	row := conn.QueryRow(getTimeoutContext(), "select mainId, chainId from emails where event=$1", ids.Event)

	if err := row.Scan(
		&ids.Main,
		&ids.Chain,
	); err != nil {
		return EmailIds{}, fmt.Errorf("Failed to get email ids for event: %w", err)
	}

	return ids, nil
}

func SetEmailChainIDs(conn *pgx.Conn, ids EmailIds) error {
	if ids.Event != 1 && ids.Event != 2 {
		return ErrUnknownEvent
	}

	_, err := conn.Exec(getTimeoutContext(), "update emails set mainId=$1, chainId=$2 where event=$3", ids.Main, ids.Chain, ids.Event)

	return err
}
