package hosting

import "github.com/pkg/errors"

type Domain struct {
	platform Platform
	ID       string
	Name     string
	Data     string
}

func (domain *Domain) Assign(server *Server) (err error) {
	err = domain.platform.AssignDomain(domain, server)
	return errors.Wrap(err, "failed to assign domain")
}

func (domain *Domain) Destroy() error {
	return domain.platform.DestroyDomain(domain.ID)
}
