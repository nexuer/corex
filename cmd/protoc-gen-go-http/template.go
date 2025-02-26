package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed http.tmpl
var httpTemplate string

type serviceDesc struct {
	ServiceName     string
	ServiceFullName string
	Metadata        string
	Methods         []*methodDesc
	MethodSets      map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name         string
	OriginalName string
	Num          int
	Request      string
	Reply        string
	Comment      string
	// http_rule
	Path         string
	Method       string
	HasVars      bool
	HasBody      bool
	Body         string
	ResponseBody string
}

func (s *serviceDesc) template() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
