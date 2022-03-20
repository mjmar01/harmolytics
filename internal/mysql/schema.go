package mysql

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/mjmar01/harmolytics/internal/log"
	"text/template"
)

//go:embed queries/init_schema.tmpl
var initSchemaQ string

type Addr struct {
	OneAddress string `yaml:"one-addr"`
	Type       string `yaml:"type"`
	Name       string `yaml:"name"`
}

func InitSchema(overwrite, wipe bool) (err error) {
	log.Task("Writing harmolytics schema to database", log.InfoLevel)
	if overwrite || wipe {
		log.Debug("Overwrite is enabled. Cleaning profile")
		err = RunQuery(fmt.Sprintf("drop schema if exists harmolytics_profile_%s", prfl))
		if err != nil {
			return
		}
	}
	if wipe {
		log.Warn("Wipe is enabled. Cleaning everything")
		err = RunQuery("drop schema if exists harmolytics_historic")
		if err != nil {
			return
		}
		err = RunQuery("drop schema if exists harmolytics_default")
		if err != nil {
			return
		}
	}
	var buf bytes.Buffer
	t, err := template.New("initSchema").Parse(initSchemaQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, prfl)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = runTemplate(buf.String())
	if err != nil {
		return
	}
	log.Done()
	return
}
