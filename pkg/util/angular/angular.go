package angular

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	Header Header      `json:"header,omitempty"`
	Body   interface{} `json:"body,omitempty"`
}

// CommitMessage construct a commit message.
// ref：https://www.ruanyifeng.com/blog/2016/01/commit_message_change_log.html
func CommitMessage(scope string, subject Subject, body interface{}) string {
	msg := Message{
		Header: Header{
			Kind:    change,
			Scope:   scope,
			Subject: subject,
		},
		Body: body,
	}
	return msg.String()
}

func (m Message) String() string {
	data, _ := json.Marshal(m)

	var buffer bytes.Buffer
	_ = json.Compact(&buffer, data)

	return fmt.Sprintf("%v\n\n%v", m.Header, buffer.String())
}

// Header ...
type Header struct {
	Kind    kind    `json:"kind,omitempty"`
	Scope   string  `json:"scope,omitempty"`
	Subject Subject `json:"subject,omitempty"`
}

func (h Header) String() string {
	return fmt.Sprintf("%v(%v): %v", h.Kind, h.Scope, h.Subject)
}

type kind = string

const (
	change kind = "change"
)

type Subject struct {
	// Operator the operator of this operation
	Operator string `json:"operator,omitempty"`

	// Date the date of commit. Do not have to specific this value when creating a commit.
	// When getting a commit, this value will fill by the commit metadata.
	Date *time.Time `json:"date,omitempty"`

	// Action the action
	Action string `json:"action,omitempty"`

	// Application the name of application.
	Application *string `json:"application,omitempty"`

	// ApplicationInstance the name of applicationInstance.
	ApplicationInstance *string `json:"applicationInstance,omitempty"`
}

func (s Subject) String() string {
	if s.ApplicationInstance != nil {
		return fmt.Sprintf("%s %s %s", s.Operator, s.Action, *s.ApplicationInstance)
	} else if s.Application != nil {
		return fmt.Sprintf("%s %s %s", s.Operator, s.Action, *s.Application)
	}
	return fmt.Sprintf("%s %s", s.Operator, s.Action)
}

func StringPtr(s string) *string {
	return &s
}
