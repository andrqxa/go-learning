package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/irlndts/go-learning/proto/todo"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "missing subcommand list or add")
		os.Exit(1)
	}

	var err error
	switch cmd := flag.Arg(0); cmd {
	case "list":
		err = list()
	case "add":
		err = add(strings.Join(flag.Args()[1:], " "))
	default:
		err = fmt.Errorf("unknown subcommand: %s", cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}

const dbPath = "mydb.pb"

func add(text string) error {
	task := &todo.Task{
		Text: text,
		Done: false,
	}

	b, err := proto.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to encode task: %s", err)
	}

	f, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open faild %s: %s", dbPath, err)
	}

	if err := gob.NewEncoder(f).Encode(int64(len(b))); err != nil {
		return fmt.Errorf("failed to encode length of message: %s", err)
	}
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("failed to write to a file %s: %s", dbPath, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %s", err)
	}

	return nil
}

func list() error {
	b, err := ioutil.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read from file: %s", err)
	}

	for {
		if len(b) == 0 {
			return nil
		} else if len(b) < 4 {
			return fmt.Errorf("remaining odd %d bytes", len(b))
		}

		var length int64
		if err := gob.NewDecoder(bytes.NewReader(b[:4])).Decode(&length); err != nil {
			return fmt.Errorf("failed to decode lentgh: %s", err)
		}

		b = b[4:]
		var task todo.Task
		if err := proto.Unmarshal(b[:length], &task); err != nil {
			return fmt.Errorf("failed to unmarshal: %s", err)
		}
		b = b[length:]

		if task.Done {
			fmt.Printf("done %s\n", task.Text)
		} else {
			fmt.Printf("todo %s\n", task.Text)
		}
	}
}
