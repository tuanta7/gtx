package netrc

import "fmt"

type Machine struct {
	Name     string
	Login    string
	Password string
}

func (m *Machine) Format() []byte {
	formattedMachine := fmt.Sprintf(""+
		"machine %s\n"+
		"login %s\n"+
		"password %s\n",
		m.Name, m.Login, m.Password,
	)

	return []byte(formattedMachine)
}
