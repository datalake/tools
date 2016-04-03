
// Copyright 2016 The Data Lake Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/peterh/liner"

	"github.com/datalake/tools/internal/config"
)

func trace(s string) (string, time.Time) {
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()

	fmt.Printf(s, float64(endTime.UnixNano()-startTime.UnixNano())/float64(1E6))
}

func Run(query string) {
	nResults := 0
	startTrace, startTime := trace("Elapsed time: %g ms\n\n")
	defer func() {
		if nResults > 0 {
			un(startTrace, startTime)
		}
	}()
	fmt.Printf("\n")
}

const (
	ps1 = "dl> "
	ps2 = "...     "

	history = ".datalake_history"
)

func Repl(cfg *config.Config) error {

	term, err := terminal(history)
	if os.IsNotExist(err) {
		fmt.Printf("creating new history file: %q\n", history)
	}
	defer persist(term, history)

	var (
		prompt = ps1

		code string
	)

	for {
		if len(code) == 0 {
			prompt = ps1
		} else {
			prompt = ps2
		}
		line, err := term.Prompt(prompt)
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				return nil
			}
			return err
		}

		term.AppendHistory(line)

		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

	}
}

// Splits a line into a command and its arguments
// e.g. ":a b c d ." will be split into ":a" and " b c d ."
func splitLine(line string) (string, string) {
	var command, arguments string

	line = strings.TrimSpace(line)

	// An empty line/a line consisting of whitespace contains neither command nor arguments
	if len(line) > 0 {
		command = strings.Fields(line)[0]

		// A line containing only a command has no arguments
		if len(line) > len(command) {
			arguments = line[len(command):]
		}
	}

	return command, arguments
}

func terminal(path string) (*liner.State, error) {
	term := liner.NewLiner()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c

		err := persist(term, history)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to properly clean up terminal: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	f, err := os.Open(path)
	if err != nil {
		return term, err
	}
	defer f.Close()
	_, err = term.ReadHistory(f)
	return term, err
}

func persist(term *liner.State, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("could not open %q to append history: %v", path, err)
	}
	defer f.Close()
	_, err = term.WriteHistory(f)
	if err != nil {
		return fmt.Errorf("could not write history to %q: %v", path, err)
	}
	return term.Close()
}