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

// RemoveDNSRecord removes the DNS A record for the given subdomain
func (s *Server) RemoveDNSRecord(sub, domain string) error {
	ctx := context.Background()
	
	// List all records for the domain
	records, _, err := s.client.Domains.Records(ctx, domain, &godo.ListOptions{})
	if err != nil {
		log.Printf("Failed to list DNS records for %s: %v", domain, err)
		return err
	}
	
	// Find and delete the A record for this subdomain
	for _, record := range records {
		if record.Type == "A" && record.Name == sub {
			_, err := s.client.Domains.DeleteRecord(ctx, domain, record.ID)
			if err != nil {
				log.Printf("Failed to delete DNS record %s.%s: %v", sub, domain, err)
				return err
			}
			log.Printf("Deleted A record for domain: %s.%s", sub, domain)
			return nil
		}
	}
	
	log.Printf("No DNS record found for %s.%s to delete", sub, domain)
	return nil
}
