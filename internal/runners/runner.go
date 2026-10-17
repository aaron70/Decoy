package runners

import "context"

type RunnerConfig interface {
	Validate() error
}

type RunnerOutput interface {
	String() string
}

type runner[C RunnerConfig, O RunnerOutput] interface {
	Run(ctx context.Context, config C) (O, error)
}

type Runner interface {
	Run(ctx context.Context, config string) (any, error)
}

type runnerWrapper[C RunnerConfig, O RunnerOutput] struct {
	Runner             runner[C, O]
	ConfigDeserializer func(string) (C, error)
}

func (w runnerWrapper[C, O]) Run(ctx context.Context, config string) (any, error) {
	c, err := w.ConfigDeserializer(config)
	if err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	res, err := w.Runner.Run(ctx, c)
	if err != nil {
		return nil, err
	}

	return res, nil
}
