package decoy

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/cespare/xxhash/v2"
)

type ParseTemplateOption func(*parseTemplateConfig) error

type parseTemplateConfig struct {
	Name            string
	Funcs           template.FuncMap
	NestedTemplates map[string]string
	Data            any
}

func WithName(name string) ParseTemplateOption {
	return func(ptc *parseTemplateConfig) error {
		ptc.Name = name
		return nil
	}
}

func WithData(data any) ParseTemplateOption {
	return func(ptc *parseTemplateConfig) error {
		ptc.Data = data
		return nil
	}
}

func WithFuncMap(funcs template.FuncMap) ParseTemplateOption {
	return func(ptc *parseTemplateConfig) error {
		ptc.Funcs = funcs
		return nil
	}
}

func WithExtraTemplate(name, tmpl string) ParseTemplateOption {
	return func(ptc *parseTemplateConfig) error {
		if ptc.NestedTemplates == nil {
			ptc.NestedTemplates = make(map[string]string)
		}
		ptc.NestedTemplates[name] = tmpl
		return nil
	}
}

func WithExtraTemplates(extraTemplates map[string]string) ParseTemplateOption {
	return func(ptc *parseTemplateConfig) error {
		ptc.NestedTemplates = extraTemplates
		return nil
	}
}

func applyConfigOptions(config *parseTemplateConfig, options ...ParseTemplateOption) error {
	for _, option := range options {
		if err := option(config); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoy) DefaultTemplateFuncMaps() template.FuncMap {
	funcMap := template.FuncMap{
		"randomInt":             d.RandomInt,
		"randomFloat":           d.RandomFloat,
		"randomBoolean":         d.RandomBoolean,
		"randomChoice":          d.RandomChoice[any],
		"randomChoiceList":      d.RandomChoiceList[any],
		"randomText":            d.RandomText,
		"randomName":            d.RandomName,
		"randomLastName":        d.RandomLastName,
		"randomFullName":        d.RandomFullName,
		"probability":           d.Probability,
		"nextIncrementalInt":    d.NextIncrementalInt,
		"setIncrementalInt":     d.SetIncrementalInt,
		"deleteIncrementalInt":   d.DeleteIncrementalInt,
		"readFileString": d.ReadFileString,
		"readFileBytes":  d.ReadFileBytes,
		"base64":         d.Base64,
		// "paginationPage":        d.PaginationOfPage,
		// "paginationSkip":        d.PaginationSkip,
		// "paginationToken":       d.PaginationNextToken,
	}
	maps.Insert(funcMap, maps.All(sprig.FuncMap()))
	return funcMap
}

func (d *Decoy) compileTemplate(tmpl string, config *parseTemplateConfig) (*template.Template, func(), error) {
	checksum := xxhash.Sum64String(fmt.Sprintf("%s%+v%+v", tmpl, config.NestedTemplates, config.Funcs))
	cleanup := func() {
		d.templatesCache.Delete(checksum)
	}
	cacheHit, exists := d.templatesCache.Load(checksum)
	if exists {
		return cacheHit.(*template.Template), cleanup, nil
	}

	compiledTemplate, err := template.
		New(config.Name).
		Funcs(config.Funcs).
		Parse(tmpl)
	if err != nil {
		return nil, cleanup, err
	}

	for name, tmpl := range config.NestedTemplates {
		_, err = compiledTemplate.New(name).Parse(tmpl)
		if err != nil {
			return nil, cleanup, err
		}
	}

	d.templatesCache.Store(checksum, compiledTemplate)
	return compiledTemplate, cleanup, nil
}

func (d *Decoy) CompileTemplate(tmpl string, options ...ParseTemplateOption) (*template.Template, func(), error) {
	config := &parseTemplateConfig{
		Funcs: d.DefaultTemplateFuncMaps(),
	}
	err := applyConfigOptions(config, options...)
	if err != nil {
		return nil, nil, err
	}
	return d.compileTemplate(tmpl, config)
}

func (d *Decoy) ParseTemplate(w io.Writer, tmpl string, options ...ParseTemplateOption) (func(), error) {
	config := &parseTemplateConfig{
		Funcs: d.DefaultTemplateFuncMaps(),
	}
	err := applyConfigOptions(config, options...)
	if err != nil {
		return nil, err
	}

	compiledTemplate, cleanTemplate, err := d.compileTemplate(tmpl, config)
	if err != nil {
		return cleanTemplate, err
	}

	return cleanTemplate, compiledTemplate.Execute(w, config.Data)
}

func (d *Decoy) ParseTemplateString(tmpl string, options ...ParseTemplateOption) (string, func(), error) {
	str := bytes.NewBufferString("")
	cleanup, err := d.ParseTemplate(str, tmpl, options...)
	if err != nil {
		return "", cleanup, err
	}
	return str.String(), cleanup, nil
}
