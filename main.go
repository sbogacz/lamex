package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultCommandFileName is the file that lamex will read for the commands that
	// you want to run. TODO: Maybe make configurable ?
	DefaultCommandFileName = "commands.txt"
)

var commands []string

func parseCommandFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open command file")
	}
	defer f.Close()

	lines := make([]string, 0, 10)
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	if err := s.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to parse command file")
	}

	return lines, nil
}

func main() {
	var err error
	commands, err = parseCommandFile(DefaultCommandFileName)
	if err != nil {
		logrus.WithError(err).Fatalf("failed to parse command file %s", DefaultCommandFileName)
	}

	lambda.Start(Handler)
}

// Handler runs the commands sequentially and will return an error
// if any of the commands error
func Handler(ctx context.Context) error {

	for _, cmd := range commands {
		if err := run(cmd); err != nil {
			return errors.Wrapf(err, "failed to run command: %s", cmd)
		}
	}

	return nil
}

func run(command string) error {
	commandParts := strings.Fields(command)
	// nothing to run
	if len(commandParts) < 1 {
		return nil
	}

	cmd := exec.Command(commandParts[0], commandParts[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrapf(err, "failed to pipe stdout for command: %s", command)
	}
	defer stdout.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		io.Copy(os.Stdout, stdout)
	}()

	if err = cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to run command: %s", command)
	}

	wg.Wait()
	return nil
}
