package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"github.com/aaron70/decoy/internal/model"
	"github.com/aaron70/decoy/internal/runners"
	"github.com/aaron70/decoy/pkg/decoy"
	"github.com/aaron70/goaty/channels"
	"github.com/aaron70/goaty/concurrency"
	"github.com/aaron70/goaty/repositories"
)

type RunnerType string

const (
	CMD  RunnerType = "cmd"
	HTTP RunnerType = "http"
)

type Runner struct {
	repo    repositories.Repository[string, model.Runner]
	Runners map[RunnerType]runners.Runner
	Decoy   *decoy.Decoy
}

func NewRunner(repo repositories.Repository[string, model.Runner], decoy *decoy.Decoy) *Runner {
	return &Runner{
		repo:  repo,
		Decoy: decoy,
		Runners: map[RunnerType]runners.Runner{
			"http": runners.NewHttpRunner(),
			"cmd":  runners.NewCmdRunner(),
		},
	}
}

func (svc Runner) Save(name, contents string) (model.Runner, error) {
	return svc.repo.Save(name, model.Runner{
		Name:   name,
		Config: contents,
	})
}

func (svc Runner) Update(name, contents string) (model.Runner, error) {
	return svc.repo.Update(name, model.Runner{
		Name:   name,
		Config: contents,
	})
}

func (svc Runner) Get(name string) (model.Runner, error) {
	return svc.repo.Get(name)
}

func (svc Runner) GetAll() ([]model.Runner, error) {
	return svc.repo.GetAll()
}

func (svc Runner) Delete(id string) (model.Runner, error) {
	return svc.repo.Delete(id)
}

func (svc Runner) Run(w io.Writer, _type RunnerType, config, tmpl string, extraTemplates map[string]string, data any, n int, workers int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	defer wg.Wait()

	runner, exists := svc.Runners[_type]
	if !exists {
		return fmt.Errorf("Invalid Runner type %q, not registered", _type)
	}

	templateCompiled, err := svc.Decoy.CompileTemplate(tmpl,
		decoy.WithName("template"),
		decoy.WithExtraTemplates(extraTemplates),
	)
	if err != nil {
		return fmt.Errorf("TemplateParseError: %w", err)
	}

	runnerCompiled, err := svc.Decoy.CompileTemplate(config,
		decoy.WithName("runner"),
		decoy.WithExtraTemplates(extraTemplates),
	)
	if err != nil {
		return fmt.Errorf("RunnerConfigurationParseError: %w", err)
	}

	bufferPool := sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	}

	mapsPool := sync.Pool{
		New: func() any { return map[string]any{"goroutines": workers, "times": n, "data": data} },
	}

	templatePool, err := concurrency.NewPool(ctx, concurrency.ToWorkerFunc(func(ctx context.Context, task string) (string, error) {
		buffer := bufferPool.Get().(*bytes.Buffer)
		buffer.Reset()
		defer bufferPool.Put(buffer)

		err := templateCompiled.Execute(buffer, data)
		if err != nil {
			return "", err
		}

		runnerData := mapsPool.Get().(map[string]any)
		defer func() {
			runnerData["template"] = nil
			mapsPool.Put(runnerData)
		}()
		runnerData["template"] = buffer.String()
		buffer.Reset()

		err = runnerCompiled.Execute(buffer, runnerData)
		if err != nil {
			return "", err
		}
		return buffer.String(), nil
	}),
		concurrency.NewPoolWithMaxWorkers(min(workers, runtime.NumCPU())),
		concurrency.NewPoolWithBufferSize(n),
	)
	if err != nil {
		return err
	}

	recordsPool, err := concurrency.NewPool(ctx, concurrency.ToWorkerFunc(runner.Run),
		concurrency.NewPoolWithMaxWorkers(workers),
		concurrency.NewPoolWithBufferSize(n),
	)
	if err != nil {
		return err
	}

	errs := make(chan error, n)
	templates, errTemplates := templatePool.ResultsErr()
	records, errRecords := recordsPool.ResultsErr()
	mergedErrs := channels.Merge(ctx, n, errs, errTemplates, errRecords)

	wg.Go(func() {
		templatePool.ProduceTasks(n, nil)
		templatePool.Close()
		templatePool.Wait()
	})

	wg.Go(func() {
		recordsPool.RecvTasks(templates)
		recordsPool.Close()
		recordsPool.Wait()
	})

	wg.Go(func() {
		defer close(errs)
		for {
			record, open, err := channels.Recv(ctx, records)
			if err != nil {
				channels.Send(ctx, errs, err)
				return
			}
			if !open {
				return
			}

			r := fmt.Sprintf("%v", record)
			if strings.HasSuffix(r, "\n") {
				fmt.Fprint(w, r)
			} else {
				fmt.Fprintln(w, r)
			}
		}
	})

	for {
		err, open, errCtx := channels.Recv(ctx, mergedErrs)
		if err != nil {
			return err
		}
		if errCtx != nil {
			return errCtx
		}
		if !open {
			break
		}
	}

	return nil
}
