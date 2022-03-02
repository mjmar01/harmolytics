package mysql

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/go-errors/errors"
	"harmolytics/harmony"
	"harmolytics/harmony/address"
	"text/template"
)

const (
	tokensQuery = "SELECT address, symbol, name, decimals FROM harmolytics_default.tokens"
	tokenQuery  = "SELECT address, symbol, name, decimals FROM harmolytics_default.tokens WHERE address = '%s'"
)

//go:embed queries/fill_tokens.tmpl
var tokensQ string

// GetTokens returns all tokens from the database
func GetTokens() (tokens []harmony.Token, err error) {
	rows, err := db.Query(tokensQuery)
	defer rows.Close()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	for rows.Next() {
		var t harmony.Token
		var addr string
		err = rows.Scan(&addr, &t.Symbol, &t.Name, &t.Decimals)
		if err != nil {
			return nil, errors.Wrap(err, 0)
		}
		t.Address, err = address.New(addr)
		if err != nil {
			return
		}
		tokens = append(tokens, t)
	}
	return
}

// GetToken returns a token for a specified address
func GetToken(addr string) (token harmony.Token, err error) {
	test, err := db.Query(fmt.Sprintf(tokenQuery, addr))
	defer test.Close()
	if err != nil {
		return harmony.Token{}, errors.Wrap(err, 0)
	}
	test.Next()
	err = test.Scan(&addr, &token.Symbol, &token.Name, &token.Decimals)
	if err != nil {
		return harmony.Token{}, errors.Wrap(err, 0)
	}
	token.Address, err = address.New(addr)
	if err != nil {
		return
	}
	return
}

// SetTokens takes a list of harmony.Token and saves those to the table tokens
func SetTokens(tokens []harmony.Token) (err error) {
	var buf bytes.Buffer
	t, err := template.New("fillTokens").Parse(tokensQ)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = t.Execute(&buf, tokens)
	if err != nil {
		return errors.Wrap(err, 0)
	}
	err = RunTemplate(buf.String())
	return
}
