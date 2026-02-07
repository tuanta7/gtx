package netrc

import "os"

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
	return nil
}
