package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/csmith/centauri/certificate"
	"github.com/csmith/centauri/proxy"
	"github.com/csmith/legotapas"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/lego"
)

var (
	userDataPath         = flag.String("user-data", "user.pem", "Path to user data")
	certificateStorePath = flag.String("certificate-store", "certs.json", "Path to certificate store")
	dnsProviderName      = flag.String("dns-provider", "", "DNS provider to use for ACME DNS-01 challenges")
	acmeEmail            = flag.String("acme-email", "", "Email address for ACME account")
	acmeDirectory        = flag.String("acme-directory", lego.LEDirectoryProduction, "ACME directory to use")
	wildcardDomains      = flag.String("wildcard-domains", "", "Space separated list of wildcard domains")
)

const (
	acmeMinCertValidity       = time.Hour * 24 * 30
	acmeMinOcspValidity       = time.Hour * 24
	selfSignedMinCertValidity = time.Hour * 24 * 7
	selfSignedOcspValidity    = time.Second
)

func certProviders() (map[string]proxy.CertificateProvider, error) {
	dnsProvider, err := legotapas.CreateProvider(*dnsProviderName)
	if err != nil {
		return nil, fmt.Errorf("dns provider error: %v", err)
	}

	legoSupplier, err := certificate.NewLegoSupplier(&certificate.LegoSupplierConfig{
		Path:        *userDataPath,
		Email:       *acmeEmail,
		DirUrl:      *acmeDirectory,
		KeyType:     certcrypto.EC384,
		DnsProvider: dnsProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("certificate supplier error: %v", err)
	}

	store, err := certificate.NewStore(*certificateStorePath)
	if err != nil {
		return nil, fmt.Errorf("certificate store error: %v", err)
	}

	var wildcardConfig = strings.Split(*wildcardDomains, " ")

	return map[string]proxy.CertificateProvider{
		"lego": certificate.NewWildcardResolver(
			certificate.NewManager(store, legoSupplier, acmeMinCertValidity, acmeMinOcspValidity),
			wildcardConfig,
		),
		"selfsigned": certificate.NewWildcardResolver(
			certificate.NewManager(store, certificate.NewSelfSignedSupplier(), selfSignedMinCertValidity, selfSignedOcspValidity),
			wildcardConfig,
		),
	}, nil
}
