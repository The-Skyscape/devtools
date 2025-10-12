package infra

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

// WithLoadBalancer creates or updates a load balancer
func WithLoadBalancer(name string, opts ...LoadBalancerOption) CloudOption {
	return CloudOption{
		OnBoot: func(c *CloudProvider, v *godo.VPC) error {
			ctx := context.Background()
			lbs, _, err := c.LoadBalancers.List(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "failed to list load balancers")
			}

			// Check if LB exists
			var lb *godo.LoadBalancer
			for _, l := range lbs {
				if l.Name == name {
					lb = &l
					break
				}
			}

			if lb == nil {
				// Build LoadBalancerRequest
				req := &godo.LoadBalancerRequest{
					Name:      name,
					Algorithm: "round_robin",
					Region:    c.vpc.RegionSlug,
					VPCUUID:   c.vpc.ID,
					SizeSlug:  "lb-small",
				}

				// Call OnInit hooks
				for _, opt := range opts {
					if opt.OnInit != nil {
						if err := opt.OnInit(c, req); err != nil {
							return errors.Wrap(err, "failed to initialize load balancer")
						}
					}
				}

				// Create new load balancer
				log.Printf("Creating load balancer: %s", name)
				lb, _, err = c.LoadBalancers.Create(ctx, req)
				if err != nil {
					return errors.Wrap(err, "failed to create load balancer")
				}
			}

			// Assign to project
			if c.projectID != "" {
				resourceID := fmt.Sprintf("do:loadbalancer:%s", lb.ID)
				c.Projects.AssignResources(ctx, c.projectID, resourceID)
			}

			// Call OnBoot hooks
			for _, opt := range opts {
				if opt.OnBoot != nil {
					if err := opt.OnBoot(c, lb); err != nil {
						return errors.Wrap(err, "failed to boot load balancer")
					}
				}
			}

			return nil
		},
	}
}

type LoadBalancerOption struct {
	OnInit func(*CloudProvider, *godo.LoadBalancerRequest) error
	OnBoot func(*CloudProvider, *godo.LoadBalancer) error
}

// WithDomain configures SSL certificate and HTTPS forwarding rule for a domain
func WithDomain(domain string) LoadBalancerOption {
	return LoadBalancerOption{
		OnInit: func(c *CloudProvider, req *godo.LoadBalancerRequest) error {
			// Ensure certificate exists for domain
			certID, err := c.ensureCertificate(domain)
			if err != nil {
				return errors.Wrap(err, "failed to ensure certificate")
			}

			// Add HTTPS forwarding rule
			req.ForwardingRules = append(req.ForwardingRules, godo.ForwardingRule{
				EntryProtocol:  "https",
				EntryPort:      443,
				TargetProtocol: "http",
				TargetPort:     80,
				CertificateID:  certID,
				TlsPassthrough: false,
			})

			return nil
		},
	}
}

// WithTargetTags configures the tag for droplet selection
func WithTargetTags(tags []string) LoadBalancerOption {
	return LoadBalancerOption{
		OnInit: func(c *CloudProvider, req *godo.LoadBalancerRequest) error {
			if len(tags) > 0 {
				req.Tag = tags[0] // DigitalOcean LBs support single tag
			}
			return nil
		},
	}
}

// WithTargetPort sets the backend port
func WithTargetPort(port int) LoadBalancerOption {
	return LoadBalancerOption{
		OnInit: func(c *CloudProvider, req *godo.LoadBalancerRequest) error {
			// Update target port in forwarding rules
			for i := range req.ForwardingRules {
				req.ForwardingRules[i].TargetPort = port
			}
			return nil
		},
	}
}

// WithHealthCheck configures the health check path
func WithHealthCheck(path string) LoadBalancerOption {
	return LoadBalancerOption{
		OnInit: func(c *CloudProvider, req *godo.LoadBalancerRequest) error {
			req.HealthCheck = &godo.HealthCheck{
				Protocol:               "http",
				Port:                   80,
				Path:                   path,
				CheckIntervalSeconds:   10,
				ResponseTimeoutSeconds: 5,
				HealthyThreshold:       5,
				UnhealthyThreshold:     3,
			}
			return nil
		},
	}
}

// WithAlgorithm sets the load balancing algorithm
func WithAlgorithm(algo string) LoadBalancerOption {
	return LoadBalancerOption{
		OnInit: func(c *CloudProvider, req *godo.LoadBalancerRequest) error {
			req.Algorithm = algo
			return nil
		},
	}
}

// WithLBSize sets the load balancer size
func WithLBSize(size string) LoadBalancerOption {
	return LoadBalancerOption{
		OnInit: func(c *CloudProvider, req *godo.LoadBalancerRequest) error {
			req.SizeSlug = size
			return nil
		},
	}
}

// ensureCertificate creates or retrieves a Let's Encrypt certificate for a domain
func (c *CloudProvider) ensureCertificate(domain string) (string, error) {
	ctx := context.Background()

	// List existing certificates
	certs, _, err := c.Certificates.List(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to list certificates")
	}

	// Check if certificate exists for this domain
	for _, cert := range certs {
		if slices.Contains(cert.DNSNames, domain) {
			log.Printf("Found existing certificate for %s: %s", domain, cert.ID)
			return cert.ID, nil
		}
	}

	// Create new Let's Encrypt certificate
	log.Printf("Creating Let's Encrypt certificate for %s", domain)
	certReq := &godo.CertificateRequest{
		Name:     fmt.Sprintf("le-%s", strings.ReplaceAll(domain, ".", "-")),
		DNSNames: []string{domain},
		Type:     "lets_encrypt",
	}

	cert, _, err := c.Certificates.Create(ctx, certReq)
	if err != nil {
		return "", errors.Wrap(err, "failed to create certificate")
	}

	return cert.ID, nil
}
