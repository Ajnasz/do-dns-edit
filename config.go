package main

import "strings"

// Config stores configuration
type Config struct {
	Domain string `required:"true"`
	Token  string `required:"true"`

	RecordType string `required:"true"`
	RecordName string `required:"true"`
	RecordData string
	RecordTTL  int

	Delete bool
	Create bool
	Update bool
	Read   bool
}

// TLD returns TLD of Domain
func (config Config) TLD() string {
	domainParts := strings.Split(config.Domain, ".")
	return strings.Join(domainParts[len(domainParts)-2:], ".")
}

// SubDomain return sobdomains
func (config Config) SubDomain() string {
	domainParts := strings.Split(config.Domain, ".")
	return strings.Join(domainParts[:len(domainParts)-2], ".")
}
