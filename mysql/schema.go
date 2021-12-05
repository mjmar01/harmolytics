package mysql

import (
	"bytes"
	_ "embed"
	"github.com/go-errors/errors"
	"text/template"
)

//go:embed queries/reset_schema.tmpl
var resetSchemaQ string

//go:embed queries/fill_known_addresses.tmpl
var knownAddressesQ string

type Addr struct {
	OneAddress string `yaml:"one-addr"`
	Type       string `yaml:"type"`
	Name       string `yaml:"name"`
}

type KnownInfo struct {
	Addrs []Addr
}

func InitSchema() (err error) {
	var buf bytes.Buffer
	t, err := template.New("resetSchema").Parse(resetSchemaQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, profile)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	if err != nil {
		return
	}
	return
}

// AddKnown fills database with known addresses specified in the config
func AddKnown(info KnownInfo) (err error) {
	var buf bytes.Buffer
	t, err := template.New("knownAddresses").Parse(knownAddressesQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, info.Addrs)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	if err != nil {
		return
	}
	return
}
