package email

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/Dannie226/cd_watcher/internal/config"
	"github.com/Dannie226/cd_watcher/internal/queries"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmailClient struct {
	cfg   *config.EmailConfig
	creds sasl.Client
	conn  *pgx.Conn
}

func NewClient(cfg *config.EmailConfig, creds sasl.Client, conn *pgx.Conn) *EmailClient {
	return &EmailClient{
		creds: creds,
		conn:  conn,
		cfg:   cfg,
	}
}

func generateMessageID(domain string) string {
	timestamp := time.Now()
	i := big.Int{}
	n := big.Int{}

	i.SetInt64(timestamp.Unix())
	n.SetInt64(1_000_000_000)

	i.Mul(&i, &n)
	n.SetInt64(int64(timestamp.Nanosecond()))

	i.Add(&i, &n)

	ts := i.Text(10)

	ent := [36]byte{}

	rand.Read(ent[:])

	ent1 := base64.URLEncoding.EncodeToString(ent[:12])
	ent2 := base64.URLEncoding.EncodeToString(ent[12:])

	return fmt.Sprintf("<%s.%s.%s@%s>", ts, ent1, ent2, domain)
}

func (c *EmailClient) SendEmail(event config.EmailEvent, message string) error {
	if c == nil {
		return nil
	}

	conn, err := smtp.DialStartTLS(c.cfg.Host, nil)

	if err != nil {
		return fmt.Errorf("Failed to connect to SMTP server: %w", err)
	}

	defer (func() {
		err := conn.Quit()

		if err != nil {
			conn.Close()
		}
	})()

	err = conn.Auth(c.creds)

	if err != nil {
		return fmt.Errorf("Failed to authenticate with SMTP server: %w", err)
	}

	buf := bytes.NewBuffer(nil)

	err = conn.Mail(c.cfg.Emailer, nil)

	if err != nil {
		return fmt.Errorf("Failed to start email: %w", err)
	}

	fmt.Fprintf(buf, "From: %s\n", c.cfg.Emailer)

	hasRecipients := false
	hasNonBCC := false

	for _, r := range c.cfg.Recipients {
		if r.Events.RecievesEmail(event) {
			err := conn.Rcpt(r.Email, nil)

			if err != nil {
				return fmt.Errorf("Failed to add recipient: %w", err)
			}

			if !r.Events.IsBCC(event) {
				if hasNonBCC {
					buf.WriteString(", ")
				} else {
					buf.WriteString("To: ")
				}

				buf.WriteString(r.Email)
				hasNonBCC = true
			}

			hasRecipients = true
		}
	}

	if !hasRecipients {
		return nil
	}

	buf.WriteRune('\n')

	ids, err := queries.GetEmailChainIDs(c.conn, event)

	if err != nil {
		return fmt.Errorf("Failed to get email chain ids: %w", err)
	}

	id := generateMessageID(c.cfg.MsgIDDomain)

	fmt.Fprintf(buf, "Message-ID: %s\n", id)

	if event.RecievesEmail(config.StartEvent) {
		if ids.Main.Valid {
			fmt.Fprintf(buf, "In-Reply-To: %s\nReferences: %s\n", ids.Main.String, ids.Main.String)
		} else {
			ids.Main = pgtype.Text{
				String: id,
				Valid:  true,
			}
		}
	} else if event.RecievesEmail(config.FinishEvent) {
		if !ids.Chain.Valid {
			return fmt.Errorf("Sending finish event without start event")
		}

		fmt.Fprintf(buf, "In-Reply-To: %s\nReferences: %s\n", ids.Chain.String, ids.Chain.String)
	}

	ids.Chain = pgtype.Text{
		String: id,
		Valid:  true,
	}

	err = queries.SetEmailChainIDs(c.conn, ids)

	if err != nil {
		return fmt.Errorf("Failed to update chain ids: %w", err)
	}

	fmt.Fprintf(buf, "Subject: %s\n\n", event.SubjectLine())

	buf.WriteString(message)

	w, err := conn.Data()

	if err != nil {
		return fmt.Errorf("Failed to get data writer: %w", err)
	}

	_, err = buf.WriteTo(w)

	if err != nil {
		return fmt.Errorf("Failed to write email: %w", err)
	}

	err = w.Close()

	if err != nil {
		return fmt.Errorf("Failed to close email writer: %w", err)
	}

	return nil
}
