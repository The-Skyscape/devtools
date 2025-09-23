package hosting

import "github.com/pkg/errors"

type Domain struct {
	Platform Platform
	ID       string
	Sub      string
	Name     string
	Type     string
	Data     string
}

func (domain *Domain) Assign(server *Server) (err error) {
	err = domain.Platform.AssignDomain(server, domain)
	return errors.Wrap(err, "failed to assign domain")
}

func (domain *Domain) Destroy() error {
	return domain.Platform.DestroyDomain(domain.ID)
}
