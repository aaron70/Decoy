package decoy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math/rand/v2"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/aaron70/decoy/internal/random"
	"github.com/aaron70/goaty/cache"
	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/validations"
	"github.com/cespare/xxhash/v2"
)

const (
	DECOY_URL_SCHEME           = "decoy"
	QUERY_PARAM_EXTRA_TEMPLATE = "extraTemplate"
	QUERY_PARAM_VALUE          = "value"
)

type Decoy struct {
	Rand            *rand.Rand
	randomText      *random.RandomText
	intIncrementals map[string]*atomic.Int64
	templates       cache.Cache[uint64, *template.Template]
	m               sync.RWMutex
}

func NewDecoy(source rand.Source) (*Decoy, error) {
	rand := rand.New(source)
	templatesCache, err := cache.NewTTLCache[uint64, *template.Template](time.Minute * 2, time.Minute)
	if err != nil {
		return nil, err
	}
	d := &Decoy{
		Rand:            rand,
		randomText:      random.NewRandomText(rand, 6),
		intIncrementals: make(map[string]*atomic.Int64),
		templates:       templatesCache,
	}
	d.LoadDefaultNgrams()
	return d, nil
}

func NewDecoyWithSeed(seed uint64) (*Decoy, error) {
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	return NewDecoy(rand.NewPCG(seed, 1))
}

func (d *Decoy) Close() {
	d.templates.Close()
}

func (d *Decoy) LoadDefaultNgrams() {
	d.randomText.NgramsFromString(random.DefaultRandomTextCorpus)
}

func (d *Decoy) LoadNgramsString(corpus string) {
	d.randomText.NgramsFromString(corpus)
}

func (d *Decoy) randomIntUnsafe(min, max int) int {
	return d.Rand.IntN(max-min) + min
}

func (d *Decoy) randomFloatUnsafe(min, max float64) float64 {
	return min + d.Rand.Float64()*(max-min)
}

func (d *Decoy) randomBooleanUnsafe() bool {
	return d.Rand.IntN(2) == 0
}

func (d *Decoy) RandomInt(min, max int) (int, error) {
	if min >= max {
		return 0, fmt.Errorf("min (%d) must be less than max (%d)", min, max)
	}
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomIntUnsafe(min, max), nil
}

func (d *Decoy) RandomFloat(min, max float64) (float64, error) {
	if min > max {
		return 0, fmt.Errorf("RandomFloat: min (%f) must be less than or equal to max (%f)", min, max)
	}
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomFloatUnsafe(min, max), nil
}

func (d *Decoy) RandomBoolean() bool {
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomBooleanUnsafe()
}

func (d *Decoy) RandomText(maxWords int) (string, error) {
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomText.RandomText(maxWords)
}

func (d *Decoy) RandomName() string {
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomText.RandomName()
}

func (d *Decoy) RandomLastName() string {
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomText.RandomLastName()
}

func (d *Decoy) RandomFullName(middleNameProbability float64) string {
	d.m.Lock()
	defer d.m.Unlock()
	name := d.randomText.RandomName()
	lastName := d.randomText.RandomLastName()
	var middle bool
	if middleNameProbability > 0 {
		if middleNameProbability >= 1 {
			middle = true
		} else {
			middle = d.randomFloatUnsafe(0, 1) < middleNameProbability
		}
	}
	if middle {
		return fmt.Sprintf("%s %s %s", name, d.randomText.RandomLastName(), lastName)
	}
	return fmt.Sprintf("%s %s", name, lastName)
}

func (d *Decoy) Probability(probability float64) bool {
	if probability <= 0 {
		return false
	}
	if probability >= 1 {
		return true
	}
	d.m.Lock()
	defer d.m.Unlock()
	return d.randomFloatUnsafe(0, 1) < probability
}

func (d *Decoy) SetIncrementalInt(id string, value int64) int64 {
	d.m.RLock()
	defer d.m.RUnlock()
	incremental, exists := d.intIncrementals[id]
	if !exists {
		incremental = &atomic.Int64{}
		d.intIncrementals[id] = incremental
	}
	incremental.Store(value)
	return value
}

func (d *Decoy) UnsetIncrementalInt(id string) string {
	d.m.RLock()
	defer d.m.RUnlock()
	delete(d.intIncrementals, id)
	return ""
}

func (d *Decoy) CurrentIncrementalInt(id string, defaultVal int64) int64 {
	d.m.RLock()
	defer d.m.RUnlock()
	incremental, exists := d.intIncrementals[id]
	if !exists {
		return defaultVal
	}
	return incremental.Load()
}

func (d *Decoy) NextIncrementalInt(id string, start, step int64) int64 {
	d.m.Lock()
	defer d.m.Unlock()
	incremental, exists := d.intIncrementals[id]
	if !exists {
		incremental = &atomic.Int64{}
		incremental.Store(start)
		d.intIncrementals[id] = incremental
	} else {
		d.intIncrementals[id].Add(step)
	}
	return incremental.Load()
}

func (d *Decoy) RandomChoiceAny(choices ...any) (any, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("requires at least one argument")
	}
	d.m.Lock()
	defer d.m.Unlock()
	return RandomChoice(d, choices...)
}

func (d *Decoy) RandomChoiceList(choices []any) (any, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("requires at least one argument")
	}
	d.m.Lock()
	defer d.m.Unlock()
	return RandomChoice(d, choices...)
}

func (d *Decoy) ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (d *Decoy) ReadFileBase64(path string) (string, error) {
	bytes, err := d.ReadFileBytes(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (d *Decoy) ReadFileString(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	return string(bytes), err
}

func (d *Decoy) JsonUnmarshalBytes(data []byte) (map[string]any, error) {
	var res map[string]any
	err := json.Unmarshal(data, &res)
	return res, err
}

func (d *Decoy) JsonUnmarshalString(data string) (map[string]any, error) {
	return d.JsonUnmarshalBytes([]byte(data))
}

func (d *Decoy) PaginationOfPage(page, size, total int) (map[string]any, error) {
	if size <= 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "size must be greater than 0, got %d", size)
	}
	if total < 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "total must be greater than or equal to 0, got %d", total)
	}

	totalPages := (total + size - 1) / size

	if page < 0 {
		page = totalPages + page
		if page < 0 {
			page *= -1
		}
	}

	length := 0
	if page >= 0 && page < totalPages {
		length = min(size, total-page*size)
	}

	ids := make([]int, length)
	id := page * size
	for i := range length {
		ids[i] = id
		id++
	}

	return map[string]any{
		"ids":   ids,
		"page":  page,
		"size":  size,
		"total": total,
	}, nil
}

func (d *Decoy) PaginationSkip(skip, limit, total int) (map[string]any, error) {
	if skip < 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "skip must be greater than or equal to 0, got %d", skip)
	}
	if limit <= 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "limit must be greater than 0, got %d", limit)
	}
	if total < 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "total must be greater than or equal to 0, got %d", total)
	}

	length := min(limit, max(0, total-skip))
	ids := make([]int, length)
	for i := range length {
		ids[i] = skip + i
	}

	return map[string]any{
		"ids":   ids,
		"skip":  skip,
		"limit": limit,
		"total": total,
	}, nil
}

func (d *Decoy) PaginationNextToken(token string, limit, total int) (map[string]any, error) {
	skip := 0
	if token != "" {
		b, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return nil, errors.NewError(errors.ErrInvalidArgument, nil, "invalid token")
		}
		skip, err = strconv.Atoi(string(b))
		if err != nil {
			return nil, errors.NewError(errors.ErrInvalidArgument, nil, "invalid token")
		}
	}
	if skip < 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "token offset must be non-negative, got %d", skip)
	}
	if limit <= 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "limit must be greater than 0, got %d", limit)
	}
	if total < 0 {
		return nil, errors.NewError(errors.ErrInvalidArgument, nil, "total must be greater than or equal to 0, got %d", total)
	}

	length := min(limit, max(0, total-skip))
	ids := make([]int, length)
	for i := range length {
		ids[i] = skip + i
	}

	nextSkip := skip + limit
	var nextToken string
	if nextSkip < total {
		nextToken = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(nextSkip)))
	}

	return map[string]any{
		"ids":       ids,
		"nextToken": nextToken,
		"limit":     limit,
		"total":     total,
	}, nil
}

type ParseTemplateOption func(*parseTemplateConfig) error

type parseTemplateConfig struct {
	Name      string
	Funcs     template.FuncMap
	Templates map[string]string
	Data      any
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
		if ptc.Templates == nil {
			ptc.Templates = make(map[string]string)
		}
		ptc.Templates[name] = tmpl
		return nil
	}
}

func WithExtraTemplates(extraTemplates map[string]string) ParseTemplateOption {
	return func(ptc *parseTemplateConfig) error {
		ptc.Templates = extraTemplates
		return nil
	}
}

func (d *Decoy) DefaultTemplateFuncMaps() template.FuncMap {
	funcMap := template.FuncMap{
		"randomInt":             d.RandomInt,
		"randomFloat":           d.RandomFloat,
		"randomBoolean":         d.RandomBoolean,
		"randomChoice":          d.RandomChoiceAny,
		"randomChoiceList":      d.RandomChoiceList,
		"randomText":            d.RandomText,
		"randomName":            d.RandomName,
		"randomLastName":        d.RandomLastName,
		"randomFullName":        d.RandomFullName,
		"probability":           d.Probability,
		"nextIncrementalInt":    d.NextIncrementalInt,
		"currentIncrementalInt": d.CurrentIncrementalInt,
		"setIncrementalInt":     d.SetIncrementalInt,
		"unsetIncrementalInt":   d.UnsetIncrementalInt,
		"readFileString":        d.ReadFileString,
		"readFileBytes":         d.ReadFileBytes,
		"readFileBase64":        d.ReadFileBase64,
		"paginationPage":        d.PaginationOfPage,
		"paginationSkip":        d.PaginationSkip,
		"paginationToken":       d.PaginationNextToken,
	}
	maps.Insert(funcMap, maps.All(sprig.FuncMap()))
	return funcMap
}

func (d *Decoy) compileTemplate(tmpl string, config *parseTemplateConfig) (*template.Template, error) {
	t, err := template.New(config.Name).
		Funcs(config.Funcs).
		Parse(tmpl)
	if err != nil {
		return nil, err
	}

	for name, tmpl := range config.Templates {
		_, err = t.New(name).Parse(tmpl)
		if err != nil {
			return nil, err
		}
	}

	key := xxhash.Sum64String(fmt.Sprintf("%s%+v",tmpl,config))
	_ = d.templates.Put(key, t)
	return t, nil
}

func (d *Decoy) CompileTemplate(tmpl string, options ...ParseTemplateOption) (*template.Template, error) {
	config := &parseTemplateConfig{
		Funcs: d.DefaultTemplateFuncMaps(),
	}

	for _, option := range options {
		if err := option(config); err != nil {
			return nil, err
		}
	}

	return d.compileTemplate(tmpl, config)
}

func (d *Decoy) ParseTemplate(w io.Writer, tmpl string, options ...ParseTemplateOption) error {
	var (
		err      error
		template *template.Template
	)
	config := &parseTemplateConfig{
		Funcs: d.DefaultTemplateFuncMaps(),
	}

	for _, option := range options {
		if err := option(config); err != nil {
			return err
		}
	}

	key := xxhash.Sum64String(fmt.Sprintf("%s%+v",tmpl,config))
	template, _ = d.templates.Get(key)
	if template == nil {
		template, err = d.compileTemplate(tmpl, config)
		if err != nil {
			return err
		}
	}

	return template.Execute(w, config.Data)
}

func (d *Decoy) ParseTemplateString(tmpl string, options ...ParseTemplateOption) (string, error) {
	str := bytes.NewBufferString("")
	err := d.ParseTemplate(str, tmpl, options...)
	if err != nil {
		return "", err
	}
	return str.String(), nil
}

type TemplateURLReferenced struct {
	Tmpl           string
	ExtraTemplates map[string]string
	Data           map[string]any
}

func (d *Decoy) ResolveTemplateURL(resolver func(string) (string, error), rawUrl string) (TemplateURLReferenced, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return TemplateURLReferenced{}, err
	}

	if url.Scheme != DECOY_URL_SCHEME {
		return TemplateURLReferenced{}, errors.NewError(errors.ErrInvalidArgument, nil, "Invalid scheme expected %q, got: %s", DECOY_URL_SCHEME, url.Scheme)
	}

	templateName := strings.TrimLeft(url.Path, "/")
	if validations.StrIsBlank(templateName) {
		return TemplateURLReferenced{}, errors.NewError(errors.ErrInvalidArgument, nil, "The template name can't be blank")
	}

	tmpl, err := resolver(templateName)
	if err != nil {
		return TemplateURLReferenced{}, errors.NewError(nil, err, "Couldn't resolve the template")
	}

	extraTemplatesNames := strings.Split(url.Query().Get(QUERY_PARAM_EXTRA_TEMPLATE), ",")
	extraTemplates := make(map[string]string, len(extraTemplatesNames))
	for _, extraTemplateName := range extraTemplatesNames {
		if validations.StrIsBlank(extraTemplateName) { continue }
		extraTemplate, err := resolver(extraTemplateName)
		if err != nil {
			return TemplateURLReferenced{}, errors.NewError(nil, err, "Couldn't resolve extra template")
		}
		extraTemplates[extraTemplateName] = extraTemplate
	}

	values := strings.Split(url.Query().Get(QUERY_PARAM_VALUE), ",")
	data := make(map[string]any)
	for _, value := range values {
		if validations.StrIsBlank(value) { continue }
		keyValue := strings.Split(value, "=") 
		if len(keyValue) != 2 {
			return TemplateURLReferenced{}, errors.NewError(errors.ErrInvalidArgument, nil, "Invalid value %q, expected format: key=value", value)
		}
		data[keyValue[0]] = keyValue[1]
	}

	return TemplateURLReferenced{
		Tmpl: tmpl,
		ExtraTemplates: extraTemplates,
		Data: data,
	}, nil
}
