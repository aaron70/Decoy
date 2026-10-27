package template

import (
	"fmt"
	"maps"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/aaron70/decoy/pkg/decoy"
	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/options"
	"github.com/cespare/xxhash/v2"
)

type NewTemplateOption func(*NewTemplateConfig) error

type NewTemplateConfig struct {
	Name            string
	Funcs           template.FuncMap
	UseDefaultFuncs bool
	NestedTemplates map[string]string
	Decoy           *decoy.Decoy
}

func WithDefaultFuncs(UseDefaultFuncs bool) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		ntc.UseDefaultFuncs = UseDefaultFuncs
		return nil
	}
}

func WithExtraFuncs(funcs template.FuncMap) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		ntc.Funcs = funcs
		return nil
	}
}

func WithDecoy(decoy *decoy.Decoy) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		if decoy == nil {
			return errors.NewError(errors.ErrNilReference, nil, "WithDecoy: The decoy instance can't be nil")
		}
		ntc.Decoy = decoy
		return nil
	}
}

type Template struct {
	*decoy.Decoy
	checksum uint64
	tmpl     *template.Template
}

func NewTemplate(tmpl string, opts ...NewTemplateOption) (*Template, error) {
	config := NewTemplateConfig{
		UseDefaultFuncs: true,
	}
	err := options.Apply(&config, opts...)
	if err != nil {
		return nil, err
	}
	if config.Decoy == nil {
		config.Decoy, err = decoy.NewDecoy()
		if err != nil {
			return nil, errors.NewError(nil, err, "Couldn't create the Decoy instance")
		}
		config.Decoy.LoadDefaultNgrams()
	}
	checksum := xxhash.Sum64String(fmt.Sprintf("%s%s%+v%+v%t", config.Name, tmpl, config.NestedTemplates, config.Funcs, config.UseDefaultFuncs))

	t := &Template{
		Decoy:    config.Decoy,
		checksum: checksum,
	}

	compiledTemplate := template.New(config.Name).
		Funcs(config.Funcs)

	if config.UseDefaultFuncs {
		compiledTemplate.Funcs(t.defaultTemplateFuncMaps())
	}

	compiledTemplate, err = compiledTemplate.Parse(tmpl)
	if err != nil {
		return nil, errors.NewError(nil, err, "Coudn't create the template")
	}

	for name, nestedTmpl := range config.NestedTemplates {
		_, err = compiledTemplate.New(name).Parse(nestedTmpl)
		if err != nil {
			return nil, errors.NewError(nil, err, "Couldn't parse nested template %q", name)
		}
	}

	t.tmpl = compiledTemplate
	return t, nil
}

func (t *Template) defaultTemplateFuncMaps() template.FuncMap {
	funcMap := template.FuncMap{
		"randomInt":            t.RandomInt,
		"randomFloat":          t.RandomFloat,
		"randomBoolean":        t.RandomBoolean,
		"randomChoice":         t.RandomChoice[any],
		"randomChoiceList":     t.RandomChoiceList[any],
		"randomText":           t.RandomText,
		"randomName":           t.RandomName,
		"randomLastName":       t.RandomLastName,
		"randomFullName":       t.RandomFullName,
		"probability":          t.Probability,
		"nextIncrementalInt":   t.NextIncrementalInt,
		"setIncrementalInt":    t.SetIncrementalInt,
		"deleteIncrementalInt": t.DeleteIncrementalInt,
		"readFileString":       t.ReadFileString,
		"readFileBytes":        t.ReadFileBytes,
		"base64":               t.Base64,
		// "paginationPage":        d.PaginationOfPage,
		// "paginationSkip":        d.PaginationSkip,
		// "paginationToken":       d.PaginationNextToken,
	}
	maps.Insert(funcMap, maps.All(sprig.FuncMap()))
	return funcMap
}
