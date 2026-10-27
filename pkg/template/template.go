package template

import (
	"encoding/binary"
	"maps"
	"slices"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/aaron70/decoy/pkg/decoy"
	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/options"
	"github.com/cespare/xxhash/v2"
)

type NewTemplateOption func(*NewTemplateConfig) error

type NewTemplateConfig struct {
	Name              string
	Funcs             template.FuncMap
	UseDefaultFuncs   bool
	LoadDefaultNgrams bool
	NestedTemplates   map[string]string
	Decoy             *decoy.Decoy
	BaseDir           string
	MaxFileSize       int64
	MissingKeyError   bool
}

func WithName(name string) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		if strings.TrimSpace(name) == "" {
			return errors.NewError(errors.ErrInvalidArgument, nil, "WithName: The template name can't be empty")
		}
		ntc.Name = name
		return nil
	}
}

func WithDefaultFuncs(useDefaultFuncs bool) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		ntc.UseDefaultFuncs = useDefaultFuncs
		return nil
	}
}

func WithDefaultNgrams(loadDefaultNgrams bool) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		ntc.LoadDefaultNgrams = loadDefaultNgrams
		return nil
	}
}

func WithExtraFuncs(funcs template.FuncMap) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		if len(funcs) == 0 {
			return nil
		}
		if ntc.Funcs == nil {
			ntc.Funcs = make(template.FuncMap, len(funcs))
		}
		maps.Copy(ntc.Funcs, funcs)
		return nil
	}
}

func WithNestedTemplates(templates map[string]string) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		if len(templates) == 0 {
			return nil
		}
		if ntc.NestedTemplates == nil {
			ntc.NestedTemplates = make(map[string]string, len(templates))
		}
		maps.Copy(ntc.NestedTemplates, templates)
		return nil
	}
}

func WithDecoy(d *decoy.Decoy) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		if d == nil {
			return errors.NewError(errors.ErrNilReference, nil, "WithDecoy: The decoy instance can't be nil")
		}
		ntc.Decoy = d
		return nil
	}
}

func WithBaseDir(dir string) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		ntc.BaseDir = dir
		return nil
	}
}

func WithMaxFileSize(size int64) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		if size < 0 {
			return errors.NewError(errors.ErrInvalidArgument, nil, "WithMaxFileSize: The maximum file size can't be negative")
		}
		ntc.MaxFileSize = size
		return nil
	}
}

func WithMissingKeyError(enabled bool) NewTemplateOption {
	return func(ntc *NewTemplateConfig) error {
		ntc.MissingKeyError = enabled
		return nil
	}
}

type Template struct {
	*decoy.Decoy
	checksum    uint64
	tmpl        *template.Template
	baseDir     string
	maxFileSize int64
}

func NewTemplate(tmpl string, opts ...NewTemplateOption) (result *Template, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = nil
			err = errors.NewError(errors.ErrInvalidArgument, nil, "NewTemplate: %v", recovered)
		}
	}()

	config := NewTemplateConfig{
		UseDefaultFuncs:   true,
		LoadDefaultNgrams: true,
		BaseDir:           ".",
	}
	if err := options.Apply(&config, opts...); err != nil {
		return nil, err
	}
	if config.Decoy == nil {
		config.Decoy, err = decoy.NewDecoy()
		if err != nil {
			return nil, errors.NewError(nil, err, "Couldn't create the Decoy instance")
		}
	}
	if config.LoadDefaultNgrams {
		config.Decoy.EnableDefaultNgrams()
	}

	t := &Template{
		Decoy:       config.Decoy,
		checksum:    computeChecksum(config.Name, tmpl, config.NestedTemplates, config.Funcs, config.UseDefaultFuncs),
		baseDir:     config.BaseDir,
		maxFileSize: config.MaxFileSize,
	}

	compiledTemplate := template.New(config.Name)
	if config.MissingKeyError {
		compiledTemplate = compiledTemplate.Option("missingkey=error")
	}
	if config.UseDefaultFuncs {
		compiledTemplate.Funcs(t.defaultTemplateFuncMaps())
	}
	compiledTemplate.Funcs(config.Funcs)

	compiledTemplate, err = compiledTemplate.Parse(tmpl)
	if err != nil {
		return nil, errors.NewError(nil, err, "Couldn't create the template")
	}

	for name, nestedTmpl := range config.NestedTemplates {
		if strings.TrimSpace(name) == "" {
			return nil, errors.NewError(errors.ErrInvalidArgument, nil, "The nested template name can't be empty")
		}
		if name == config.Name {
			return nil, errors.NewError(errors.ErrInvalidArgument, nil, "The nested template name %q conflicts with the root template name", name)
		}
		if _, err = compiledTemplate.New(name).Parse(nestedTmpl); err != nil {
			return nil, errors.NewError(nil, err, "Couldn't parse nested template %q", name)
		}
	}

	t.tmpl = compiledTemplate
	return t, nil
}

func (t *Template) Checksum() uint64 {
	return t.checksum
}

func (t *Template) defaultTemplateFuncMaps() template.FuncMap {
	funcMap := sprig.FuncMap()
	maps.Copy(funcMap, template.FuncMap{
		"randomInt":            t.RandomInt,
		"randomFloat":          t.RandomFloat,
		"randomBoolean":        t.RandomBoolean,
		"randomChoice":         t.RandomChoice[any],
		"randomChoiceList":     t.randomChoiceList,
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
		"base64String":         t.Base64String,
		"base64Decode":         t.Base64Decode,
	})
	return funcMap
}

func computeChecksum(name, tmpl string, nested map[string]string, funcs template.FuncMap, useDefaultFuncs bool) uint64 {
	hasher := xxhash.New()
	write := func(value string) {
		var size [8]byte
		binary.LittleEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = hasher.Write(size[:])
		_, _ = hasher.Write([]byte(value))
	}

	write(name)
	write(tmpl)
	for _, nestedName := range slices.Sorted(maps.Keys(nested)) {
		write(nestedName)
		write(nested[nestedName])
	}
	for _, funcName := range slices.Sorted(maps.Keys(funcs)) {
		write(funcName)
	}
	if useDefaultFuncs {
		write("1")
	} else {
		write("0")
	}
	return hasher.Sum64()
}
