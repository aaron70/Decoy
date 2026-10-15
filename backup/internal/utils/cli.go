package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
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
