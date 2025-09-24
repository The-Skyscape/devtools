package digitalocean

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
				ID:       fmt.Sprintf("%d:%s", record.ID, domain.Name), // Format: "recordID:domainName"
				Sub:      record.Name,
				Name:     domain.Name,
				Type:     record.Type,
				Data:     record.Data,
			}, nil
		}
	}

	return nil, errors.New("domain not found: " + domain.Name)
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

func (client *DigitalOceanClient) DestroyDomain(recordInfo string) error {
	parts := strings.Split(recordInfo, ":")
	if len(parts) != 2 {
		return errors.New("invalid domain record info format, expected 'recordID:domainName'")
	}

	recordID, err := strconv.Atoi(parts[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse record ID")
	}

	_, err = client.Domains.DeleteRecord(context.Background(), parts[1], recordID)
	return errors.Wrap(err, "failed to delete domain record")
}
