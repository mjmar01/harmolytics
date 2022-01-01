package mysql

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/go-errors/errors"
	"text/template"
)

//go:embed queries/init_schema.tmpl
var initSchemaQ string

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

func InitSchema(overwrite bool) (err error) {
	if overwrite {
		err = RunQuery(fmt.Sprintf("drop schema if exists harmolytics_profile_%s", profile))
		if err != nil {
			return
		}
	}
	var buf bytes.Buffer
	t, err := template.New("initSchema").Parse(initSchemaQ)
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
