package digitalocean

import (
	"context"
	"log"

	"github.com/digitalocean/godo"
)

func (s *Server) Alias(sub, domain string) (err error) {
	ctx := context.Background()
	records, _, err := s.client.Domains.Records(ctx, domain, &godo.ListOptions{})
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.Type == "A" && record.Name == sub && record.Data == s.IP {
			log.Printf("DNS record already exists for domain: %s.%s ", sub, domain)
			return nil
		}
	}

	record := &godo.DomainRecordEditRequest{Type: "A", Name: sub, Data: s.IP, TTL: 3600}
	if _, _, err = s.client.Domains.CreateRecord(ctx, domain, record); err != nil {
		return err
	}

	log.Printf("Created A record for domain: %s.%s with IP: %s", sub, domain, s.IP)
	return
}
