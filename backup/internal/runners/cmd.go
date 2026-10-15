package runners

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

type CmdRunnerConfig string

func (c *CmdRunnerConfig) Validate() error {
	if *c == CmdRunnerConfig("") {
		return fmt.Errorf("The command can't be empty")
	}
	return nil
}

type CmdRunnerOutput struct {
	Response string
}

func (o CmdRunnerOutput) String() string {
	return o.Response
}

type CmdRunner struct {
	Client exec.Cmd
}

func NewCmdRunner() Runner {
	return &runnerWrapper[*CmdRunnerConfig, *CmdRunnerOutput]{
		Runner: CmdRunner{},
		ConfigDeserializer: func(config string) (*CmdRunnerConfig, error) {
			configuration := CmdRunnerConfig(config)
			return &configuration, nil
		},
	}
}

func (r CmdRunner) Run(ctx context.Context, config *CmdRunnerConfig) (*CmdRunnerOutput, error) {
	args, err := splitCommand(string(*config))
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	outerr, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return &CmdRunnerOutput{string(outerr)}, nil
}

func splitCommand(command string) ([]string, error) {
	args := []string{}
	arg := strings.Builder{}
	strIndex := -1
	argIndex := -1
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == ' ' && strIndex < 0 {
			args = append(args, arg.String())
			arg.Reset()
			argIndex = -1
			continue
		}
		_, err := arg.WriteRune(r)
		if err != nil {
			return nil, err
		}

		if slices.Contains([]rune{'"', '\'', '`'}, r) {
			if strIndex < 0 {
				strIndex = i
				argIndex = len(arg.String()) - 1
			} else if runes[strIndex] == r {
				strIndex = -1
				argIndex = -1
			}
		}

		if i == len(runes)-1 && strIndex >= 0 {
			i = strIndex
			k := arg.String()[0:argIndex]
			arg.Reset()
			arg.WriteString(k)
			arg.WriteRune(runes[strIndex])
			strIndex = -1
		}
	}

	if arg.Len() > 0 {
		args = append(args, arg.String())
	}

	return args, nil
}
