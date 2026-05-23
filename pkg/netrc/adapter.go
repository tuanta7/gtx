package netrc

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sort"
	"strings"
)

type Adapter struct {
	file     *os.File
	machines map[string]*Machine
}

func NewAdapter(filePath string) (*Adapter, error) {
	netrcFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		file: netrcFile,
	}, nil
}

func (a *Adapter) Set(machine *Machine) error {
	if err := a.load(); err != nil {
		return err
	}

	a.machines[machine.Name] = machine
	return a.write()
}

func (a *Adapter) Get(name string) (*Machine, error) {
	if err := a.load(); err != nil {
		return nil, err
	}

	machine, ok := a.machines[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	return machine, nil
}

func (a *Adapter) load() error {
	if a.machines != nil {
		return nil
	}

	if _, err := a.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	machines, err := Parse(a.file)
	if err != nil {
		return err
	}

	a.machines = machines
	return nil
}

func (a *Adapter) write() error {
	var names []string
	for name := range a.machines {
		names = append(names, name)
	}
	sort.Strings(names)

	var content bytes.Buffer
	for i, name := range names {
		if i > 0 {
			content.WriteByte('\n')
		}
		content.Write(a.machines[name].Format())
	}

	if err := a.file.Truncate(0); err != nil {
		return err
	}
	if _, err := a.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := a.file.Write(content.Bytes()); err != nil {
		return err
	}

	return a.file.Sync()
}

func Parse(reader io.Reader) (map[string]*Machine, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)

	machines := map[string]*Machine{}
	var current *Machine

	for scanner.Scan() {
		token := scanner.Text()
		switch token {
		case "machine":
			if !scanner.Scan() {
				break
			}
			current = &Machine{Name: scanner.Text()}
			machines[current.Name] = current
		case "login":
			if current == nil || !scanner.Scan() {
				break
			}
			current.Login = scanner.Text()
		case "password":
			if current == nil || !scanner.Scan() {
				break
			}
			current.Password = scanner.Text()
		default:
			if strings.HasPrefix(token, "#") {
				continue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return machines, nil
}
