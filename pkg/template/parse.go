package template

import (
	"bytes"
	"io"

	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/options"
)

type ParseTemplateOption func(*ParseTemplateConfig) error

type ParseTemplateConfig struct {
	Data     any
	Template string
}

func WithData(data any) ParseTemplateOption {
	return func(config *ParseTemplateConfig) error {
		config.Data = data
		return nil
	}
}

func WithExecuteTemplate(name string) ParseTemplateOption {
	return func(config *ParseTemplateConfig) error {
		if name == "" {
			return errors.NewError(errors.ErrInvalidArgument, nil, "WithExecuteTemplate: The template name can't be empty")
		}
		config.Template = name
		return nil
	}
}

func (t *Template) Parse(wr io.Writer, opts ...ParseTemplateOption) error {
	config, err := options.ApplyZero(opts...)
	if err != nil {
		return err
	}
	if config.Template != "" {
		return t.tmpl.ExecuteTemplate(wr, config.Template, config.Data)
	}
	return t.tmpl.Execute(wr, config.Data)
}

func (t *Template) ParseString(opts ...ParseTemplateOption) (string, error) {
	str := bytes.NewBufferString("")
	err := t.Parse(str, opts...)
	return str.String(), err
}
