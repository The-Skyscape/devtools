package digitalocean

import (
	"context"
	"fmt"
	"strconv"

	"github.com/The-Skyscape/devtools/pkg/hosting"
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

func (client *DigitalOceanClient) LookupDomain(domain *hosting.Domain) (*hosting.Domain, error) {
	if domain == nil {
		return nil, errors.New("domain is nil")
	}

	ctx := context.Background()
	records, _, err := client.Domains.Records(ctx, domain.Name, &godo.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list domain records")
	}

	for _, record := range records {
		if record.Type == domain.Type && record.Name == domain.Sub {
			return &hosting.Domain{
				Platform: client,
				ID:       fmt.Sprintf("%d", record.ID),
				Sub:      record.Name,
				Name:     domain.Name,
				Type:     record.Type,
				Data:     record.Data,
			}, nil
		}
	}

	return nil, nil
}

func (client *DigitalOceanClient) AssignDomain(server *hosting.Server, domain *hosting.Domain) error {
	if server == nil {
		return errors.New("server is nil")
	}

	if domain == nil {
		return errors.New("domain is nil")
	}

	if existing, _ := client.LookupDomain(domain); existing != nil {
		return errors.New("a record already exists for that domain")
	}

	ctx := context.Background()
	_, _, err := client.Domains.CreateRecord(ctx, domain.Name, &godo.DomainRecordEditRequest{
		Type:     domain.Type,
		Name:     domain.Sub,
		Data:     server.IP,
		Priority: 10,
		Port:     80,
		TTL:      3600,
	})

	return errors.Wrap(err, "failed to create domain record")
}

func (client *DigitalOceanClient) DestroyDomain(id string) error {
	ctx := context.Background()
	domain, _, err := client.Domains.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to lookup domain")
	}

	domainID, err := strconv.Atoi(id)
	if err != nil {
		return errors.Wrap(err, "failed to parse domain ID")
	}

	// Delete the domain
	_, err = client.Domains.DeleteRecord(ctx, domain.Name, domainID)
	if err != nil {
		return errors.Wrap(err, "failed to delete domain")
	}

	return nil
}
