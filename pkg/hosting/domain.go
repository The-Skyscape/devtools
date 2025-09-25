package hosting

import "github.com/pkg/errors"

// Domain represents a DNS record that can point to servers.
// A Domain is created and managed by a Platform.
//
// The Platform field is set automatically when the domain is created
// via Platform.AssignDomain or retrieved via Platform.LookupDomain.
// This allows domain methods to delegate operations back to the platform.
type Domain struct {
	// Platform is the platform that created or manages this domain record.
	// This field is set automatically by Platform methods.
	Platform Platform

	// ID is the unique identifier for this domain record within the platform.
	ID string

	// Sub is the subdomain part of the record.
	// For example, "www" in "www.example.com", or "@" for the root domain.
	Sub string

	// Name is the base domain name.
	// For example, "example.com" in "www.example.com".
	Name string

	// Type is the DNS record type.
	// Common values include "A", "AAAA", "CNAME", "MX", "TXT".
	Type string

	// Data is the value of the DNS record.
	// For A records, this is typically an IP address.
	// For CNAME records, this is another domain name.
	Data string
}

// Assign creates or updates this domain record to point to the specified server.
// This is a convenience method that delegates to Platform.AssignDomain.
//
// The server's IP address will be used as the record's data for A records.
//
// Example:
//	domain := &Domain{Name: "example.com", Sub: "app", Type: "A"}
//	err := domain.Assign(server)
func (domain *Domain) Assign(server *Server) (err error) {
	err = domain.Platform.AssignDomain(server, domain)
	return errors.Wrap(err, "failed to assign domain")
}

// Destroy removes this DNS record.
// This is a convenience method that delegates to Platform.DestroyDomain.
//
// This operation cannot be undone.
//
// Example:
//	err := domain.Destroy()
func (domain *Domain) Destroy() error {
	return domain.Platform.DestroyDomain(domain.ID)
}
