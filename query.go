package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"os"
	"strings"

	"github.com/mitchellh/go-ps"
)

type Query struct {
	Type        string
	Path        string
	Check       string
	Sender      string
	Msg         string
	Id          uint64
	ResponderID int64
}

func (q *Query) Do() error {
	switch q.Type {
	case "file_contains":
		return fileContains(q.Check, q.Path)
	case "file_exists":
		return fileExists(q.Path)
	case "process_running":
		return processRunning(q.Check)
	default:
		return fmt.Errorf("query type doesn't exist")
	}

}
func fileContains(check, path string) error {

	bytz, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(bytz), check) {
		return fmt.Errorf("%s doesn't exist in file %s.", check, path)
	}

	return nil
}

func fileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func processRunning(check string) error {
	prcs, err := ps.Processes()
	if err != nil {
		return err
	}
	for _, proc := range prcs {
		if strings.Contains(proc.Executable(), check) {
			return nil
		}

	}
	return fmt.Errorf("Couldn't find process %s.", check)
}

// Basic Nostalgia
func parseQuery() *Query {
	var q Query
	host, port, err := net.SplitHostPort(*httpAddr)
	if host == "" {
		q.Sender = fmt.Sprintf("http://0.0.0.0:%s/response", port)
	} else if err == nil {
		q.Sender = fmt.Sprintf("http://%s:%s/response", host, port)
	} else {
		// Failsafe while testing
		q.Sender = fmt.Sprintf("http://0.0.0.0:8080/response", port)
	}
	reader := bufio.NewReader(os.Stdin)
START:
	fmt.Print("Enter Type (file_contains, file_exists, process_running): ")
	q.Type, _ = reader.ReadString('\n')
	q.Type = strings.TrimSpace(q.Type)
	switch q.Type {
	case "file_contains":
		break
	case "file_exists":
		break
	case "process_running":
		goto CHECK
		break
	default:
		fmt.Printf("%s is an invalid query type.\n", q.Type)
		goto START
	}

	fmt.Print("Enter path of the file: ")
	q.Path, _ = reader.ReadString('\n')
	q.Path = strings.TrimSpace(q.Path)
	if q.Type == "file_exists" {
		return &q
	}
CHECK:
	fmt.Print("Enter string you want to check for : ")
	q.Check, _ = reader.ReadString('\n')
	q.Check = strings.TrimSpace(q.Check)
	return &q
}

func response(w http.ResponseWriter, r *http.Request) {
	var q Query
	dec := gob.NewDecoder(r.Body)

	err := dec.Decode(&q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Printf("%+v\n", q)

}
