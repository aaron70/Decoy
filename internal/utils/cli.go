package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)


func AskForInput(r io.Reader, w io.Writer, msg string, args ...any) (string, error) {
	fmt.Fprintf(w, msg, args...)
	return readLine(r)
}

func readLine(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

func ReadStringFrom(r io.Reader) (string, error) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Go(func() {
		select {
		case <-ctx.Done():
		case <-time.Tick(time.Second * 2):
			fmt.Println("Waiting input from stdin...")
		}
	})

	contents, err := io.ReadAll(r)
	return string(contents), err
}

func getEditorCommand() []string {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return strings.Fields(v)
		}
	}
	if runtime.GOOS == "windows" {
		return []string{"notepad"}
	}
	return []string{"vi"}
}

func EditText(initial string) ([]byte, error) {
	f, err := os.CreateTemp("", "decoy-tmp-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(initial); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	argv := getEditorCommand()
	cmd := exec.Command(argv[0], append(argv[1:], f.Name())...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Couldn't open the editor %q, failed with: %w", argv[0], err)
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}
	return data, nil
}
