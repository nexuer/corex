package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed errors.tmpl
var errorsTemplate string

type errorInfo struct {
	Name       string
	Value      string
	HTTPCode   int
	CamelValue string
	Comment    string
	HasComment bool
}

type errorWrapper struct {
	Errors []*errorInfo
}

func (e *errorWrapper) template() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("errors").Parse(strings.TrimSpace(errorsTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, e); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
