package template

import (
	"bytes"
	"io"

	"github.com/aaron70/goaty/options"
)

type ParseTemplateOption func(*ParseTemplateConfig) error

type ParseTemplateConfig struct {
	Data any
}

func (t *Template) Parse(wr io.Writer, opts ...ParseTemplateOption) error {
	config, err := options.ApplyZero(opts...)
	if err != nil {
		return err
	}
	return t.tmpl.Execute(wr, config.Data)
}

func (t *Template) ParseString(opts ...ParseTemplateOption) (string, error) {
	str := bytes.NewBufferString("")
	err := t.Parse(str, opts...)
	return str.String(), err
}
