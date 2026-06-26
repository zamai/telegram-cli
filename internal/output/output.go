// Package output provides a small result writer shared by all tg commands.
//
// Commands produce a typed result and hand it to a Printer. In JSON mode the
// result is wrapped in a stable envelope ({"schema":N,"data":...}) so agents can
// parse it; in text mode a human-friendly rendering is used. All machine output
// goes to stdout; logs and progress must go to stderr.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-faster/errors"
)

// schemaVersion is the JSON envelope version. Bump on breaking output changes.
const schemaVersion = 1

// Format is the output format selected by the global --output flag.
type Format string

// Supported formats.
const (
	Text Format = "text"
	JSON Format = "json"
)

// Formats lists the valid output formats (for completion and validation).
func Formats() []string { return []string{string(Text), string(JSON)} }

// ParseFormat validates and returns the format for s.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case Text:
		return Text, nil
	case JSON:
		return JSON, nil
	default:
		return "", errors.Errorf("unknown output format %q (want %s)", s, strings.Join(Formats(), " or "))
	}
}

// TextMarshaler is implemented by results that render a custom text form.
// Results without it fall back to fmt's default formatting in text mode.
type TextMarshaler interface {
	MarshalText(w io.Writer) error
}

// Printer writes command results in the configured format.
type Printer struct {
	format  Format
	w       io.Writer
	account string // optional account label, included in JSON / text headers
}

// New returns a Printer writing to w in the given format.
func New(format Format, w io.Writer) *Printer {
	return &Printer{format: format, w: w}
}

// Format reports the printer's format.
func (p *Printer) Format() Format { return p.format }

// SetAccount sets the account label included with each emitted result. Empty
// disables it (single-account mode).
func (p *Printer) SetAccount(label string) { p.account = label }

type envelope struct {
	Schema  int    `json:"schema"`
	Account string `json:"account,omitempty"`
	Data    any    `json:"data"`
}

// Emit writes a single result value.
func (p *Printer) Emit(v any) error {
	if p.format == JSON {
		enc := json.NewEncoder(p.w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(envelope{Schema: schemaVersion, Account: p.account, Data: v}); err != nil {
			return errors.Wrap(err, "encode json")
		}
		return nil
	}

	if p.account != "" {
		if _, err := fmt.Fprintf(p.w, "== %s ==\n", p.account); err != nil {
			return errors.Wrap(err, "write account header")
		}
	}

	if tm, ok := v.(TextMarshaler); ok {
		return tm.MarshalText(p.w)
	}
	if _, err := fmt.Fprintln(p.w, v); err != nil {
		return errors.Wrap(err, "write text")
	}
	return nil
}

// EmitLine writes a single stream event as one line. JSON mode keeps the event
// fields at the top level and adds account when provided; text mode renders the
// event through TextMarshaler and prefixes the account on the same line.
func (p *Printer) EmitLine(account string, v any) error {
	if p.format == JSON {
		b, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "encode json line")
		}
		if account != "" {
			b, err = addAccountField(b, account)
			if err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(p.w, string(b)); err != nil {
			return errors.Wrap(err, "write json line")
		}
		return nil
	}

	line, err := textLine(v)
	if err != nil {
		return err
	}
	if account != "" {
		line = "[" + account + "] " + line
	}
	if _, err := fmt.Fprintln(p.w, line); err != nil {
		return errors.Wrap(err, "write text line")
	}
	return nil
}

func addAccountField(b []byte, account string) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, errors.Wrap(err, "decode json line")
	}
	if obj == nil {
		return nil, errors.New("json line must be an object")
	}
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return nil, errors.Wrap(err, "encode account")
	}
	obj["account"] = accountJSON
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrap(err, "encode json line")
	}
	return out, nil
}

func textLine(v any) (string, error) {
	var buf bytes.Buffer
	if tm, ok := v.(TextMarshaler); ok {
		if err := tm.MarshalText(&buf); err != nil {
			return "", err
		}
		return strings.TrimRight(buf.String(), "\n"), nil
	}
	return fmt.Sprint(v), nil
}
