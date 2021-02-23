package stepdef

import (
	"github.com/dpakach/middleware/gherkin"
	"github.com/dpakach/middleware/request"
)

type Provisioning struct {
	Suite *gherkin.Suite
	Ocs request.Ocs
}

func (p *Provisioning) Register() {
	p.Suite.Given("user {{s}} has been created with default attributes", func(user string) error {
		err := p.Ocs.CreateUser("user1", "1234", "user1@example.com", "User One")
		if err != nil  {
			return err
		}
		return nil
	})
}
