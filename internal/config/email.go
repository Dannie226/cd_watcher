package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

type loginType string
type EmailEvent int

const (
	Anonymous loginType = "anonymous"
	External  loginType = "external"
	OAuth     loginType = "oauth"
	Plain     loginType = "plain"
)

func (l *loginType) UnmarshalJSON(b []byte) error {
	var s string

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case "anonymous":
		*l = Anonymous

	case "external":
		*l = External

	case "oauth":
		*l = OAuth

	case "plain":
		*l = Plain

	default:
		return fmt.Errorf("Unknown login type: %s", s)
	}

	return nil
}

const (
	DeployStartEvent     EmailEvent = 1 << 0
	DeployFinishEvent    EmailEvent = 1 << 2
	RollbackStartEvent   EmailEvent = 1 << 4
	RollbackFinishEvent  EmailEvent = 1 << 6
	HealthCheckFailEvent EmailEvent = 1 << 8
)

type eventStruct struct {
	Name string `json:"name"`
	BCC  bool   `json:"bcc"`
}

func (e *EmailEvent) UnmarshalJSON(b []byte) error {
	var evs []eventStruct

	if err := json.Unmarshal(b, &evs); err != nil {
		return err
	}

	mask := 0

	for _, ev := range evs {
		bit := 0
		switch ev.Name {
		case "deploy_start":
			bit = int(DeployStartEvent)

		case "deploy_finish":
			bit = int(DeployFinishEvent)

		case "rollback_start":
			bit = int(RollbackStartEvent)

		case "rollback_finish":
			bit = int(RollbackFinishEvent)

		case "health_check_fail":
			bit = int(HealthCheckFailEvent)

		default:
			return fmt.Errorf("Unknown email event: %s", ev.Name)
		}

		mask |= bit

		if ev.BCC {
			mask |= bit << 1
		}
	}

	*e = EmailEvent(mask)

	return nil
}

func (e EmailEvent) RecievesEmail(ev EmailEvent) bool {
	return e&ev != 0
}

func (e EmailEvent) IsBCC(ev EmailEvent) bool {
	return e&(ev<<1) != 0
}

func (e EmailEvent) String() string {
	buf := strings.Builder{}
	buf.WriteRune('[')

	if e&DeployStartEvent != 0 {
		buf.WriteString("deploy_start")

		if e&(DeployStartEvent<<1) != 0 {
			buf.WriteString(" (BCC'd)")
		}
	}

	if e&DeployFinishEvent != 0 {
		if buf.Len() > 1 {
			buf.WriteString(", ")
		}

		buf.WriteString("deploy_finish")

		if e&(DeployFinishEvent<<1) != 0 {
			buf.WriteString(" (BCC'd)")
		}
	}

	if e&RollbackStartEvent != 0 {
		if buf.Len() > 1 {
			buf.WriteString(", ")
		}

		buf.WriteString("rollback_start")

		if e&(RollbackStartEvent<<1) != 0 {
			buf.WriteString(" (BCC'd)")
		}
	}

	if e&RollbackFinishEvent != 0 {
		if buf.Len() > 1 {
			buf.WriteString(", ")
		}

		buf.WriteString("rollback_finish")

		if e&(RollbackFinishEvent<<1) != 0 {
			buf.WriteString(" (BCC'd)")
		}
	}

	if e&HealthCheckFailEvent != 0 {
		if buf.Len() > 1 {
			buf.WriteString(", ")
		}

		buf.WriteString("health_check_fail")

		if e&(HealthCheckFailEvent<<1) != 0 {
			buf.WriteString(" (BCC'd)")
		}
	}

	buf.WriteRune(']')

	return buf.String()
}

type EmailRecipient struct {
	Email  string `json:"email"`
	Events EmailEvent
}

type EmailConfig struct {
	Host       string           `json:"host"`
	Emailer    string           `json:"emailer"`
	LoginType  loginType        `json:"login"`
	Recipients []EmailRecipient `json: "recipients"`
}
