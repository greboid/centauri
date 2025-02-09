package proxy

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCertManager struct {
	certificate       *tls.Certificate
	err               error
	preferredSupplier string
	subject           string
	altNames          []string
}

func (f *fakeCertManager) GetCertificate(preferredSupplier string, subject string, altNames []string) (*tls.Certificate, error) {
	f.preferredSupplier = preferredSupplier
	f.subject = subject
	f.altNames = altNames
	return f.certificate, f.err
}

var dummyCert = &tls.Certificate{}

func Test_Manager_SetRoutes_returnsErrorIfGetCertificateFails(t *testing.T) {
	certManager := &fakeCertManager{
		err: fmt.Errorf("ruh roh"),
	}

	manager := NewManager(certManager)
	err := manager.SetRoutes([]*Route{{
		Domains: []string{"example.com"},
	}})
	assert.Error(t, err)
	assert.Equal(t, "example.com", certManager.subject)
	assert.Equal(t, []string{}, certManager.altNames)
}

func Test_Manager_SetRoutes_returnsErrorIfDomainIsInvalid(t *testing.T) {
	manager := NewManager(nil)
	err := manager.SetRoutes([]*Route{{
		Domains: []string{"example..com"},
	}})
	assert.Error(t, err)
}

func Test_Manager_RouteForDomain_returnsNullIfNoRouteFound(t *testing.T) {
	certManager := &fakeCertManager{
		err: fmt.Errorf("ruh roh"),
	}

	manager := NewManager(certManager)
	res := manager.RouteForDomain("example.com")
	assert.Nil(t, res)
}

func Test_Manager_RouteForDomain_returnsCertificateForDomain(t *testing.T) {
	certManager := &fakeCertManager{
		certificate: dummyCert,
	}

	manager := NewManager(certManager)
	route := &Route{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}
	_ = manager.SetRoutes([]*Route{route})

	assert.Equal(t, route, manager.RouteForDomain("example.com"))
	assert.Equal(t, route, manager.RouteForDomain("test.example.com"))
	assert.Equal(t, route, manager.RouteForDomain("test.deep.example.com"))
}

func Test_Manager_CertificateForClient_returnsNullIfNoRouteFound(t *testing.T) {
	certManager := &fakeCertManager{
		err: fmt.Errorf("ruh roh"),
	}

	manager := NewManager(certManager)
	res, err := manager.CertificateForClient(&tls.ClientHelloInfo{ServerName: "example.com"})
	assert.Nil(t, res)
	assert.NoError(t, err)
}

func Test_Manager_CertificateForClient_returnsCertificateForDomain(t *testing.T) {
	certManager := &fakeCertManager{
		certificate: dummyCert,
	}

	manager := NewManager(certManager)
	route := &Route{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}
	_ = manager.SetRoutes([]*Route{route})

	res, err := manager.CertificateForClient(&tls.ClientHelloInfo{ServerName: "example.com"})
	assert.Equal(t, dummyCert, res)
	assert.NoError(t, err)

	res, err = manager.CertificateForClient(&tls.ClientHelloInfo{ServerName: "test.example.com"})
	assert.Equal(t, dummyCert, res)
	assert.NoError(t, err)

	res, err = manager.CertificateForClient(&tls.ClientHelloInfo{ServerName: "test.deep.example.com"})
	assert.Equal(t, dummyCert, res)
	assert.NoError(t, err)
}

func Test_Manager_CertificateForClient_returnsErrorIfNoProviderConfigured(t *testing.T) {
	manager := NewManager(nil)
	route := &Route{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}
	_ = manager.SetRoutes([]*Route{route})

	_, err := manager.CertificateForClient(&tls.ClientHelloInfo{ServerName: "example.com"})
	assert.Error(t, err)
}

func Test_Manager_SetRoutes_setsCertificateOnRoutes(t *testing.T) {
	certManager := &fakeCertManager{
		certificate: dummyCert,
	}

	manager := NewManager(certManager)
	_ = manager.SetRoutes([]*Route{{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}})

	assert.Equal(t, dummyCert, manager.RouteForDomain("example.com").certificate)
	assert.Equal(t, dummyCert, manager.RouteForDomain("test.example.com").certificate)
	assert.Equal(t, dummyCert, manager.RouteForDomain("test.deep.example.com").certificate)
}

func Test_Manager_SetRoutes_ignoresCertificateIfProviderNotConfigured(t *testing.T) {
	manager := NewManager(nil)
	_ = manager.SetRoutes([]*Route{{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}})

	assert.Nil(t, manager.RouteForDomain("example.com").certificate)
	assert.Nil(t, manager.RouteForDomain("test.example.com").certificate)
	assert.Nil(t, manager.RouteForDomain("test.deep.example.com").certificate)
}

func Test_Manager_SetRoutes_removesPreviousRoutes(t *testing.T) {
	certManager := &fakeCertManager{
		certificate: dummyCert,
	}

	manager := NewManager(certManager)
	_ = manager.SetRoutes([]*Route{{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}})
	_ = manager.SetRoutes([]*Route{})

	assert.Nil(t, manager.RouteForDomain("example.com"))
	assert.Nil(t, manager.RouteForDomain("test.example.com"))
	assert.Nil(t, manager.RouteForDomain("test.deep.example.com"))
}

func Test_Manager_CheckCertificates_returnsErrorIfGetCertificateFails(t *testing.T) {
	certManager := &fakeCertManager{
		certificate: dummyCert,
	}

	manager := NewManager(certManager)
	_ = manager.SetRoutes([]*Route{{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}})

	certManager.err = fmt.Errorf("ruh roh")
	err := manager.CheckCertificates()
	assert.Error(t, err)
}

func Test_Manager_CheckCertificates_passesSupplierSpecifiedByRoute(t *testing.T) {
	certManager := &fakeCertManager{}

	manager := NewManager(certManager)
	_ = manager.SetRoutes([]*Route{{
		Domains:  []string{"test.deep.example.com", "test.example.com", "example.com"},
		Provider: "f2",
	}})

	assert.Equal(t, "f2", certManager.preferredSupplier)
	assert.Equal(t, "test.deep.example.com", certManager.subject)
}

func Test_Manager_CheckCertificates_updatesAllCertificates(t *testing.T) {
	certManager := &fakeCertManager{
		certificate: dummyCert,
	}

	manager := NewManager(certManager)
	_ = manager.SetRoutes([]*Route{{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}, {
		Domains: []string{"test.example.net"},
	}})

	newCert := &tls.Certificate{OCSPStaple: []byte("Different!")}
	certManager.certificate = newCert
	err := manager.CheckCertificates()
	require.NoError(t, err)

	assert.Equal(t, newCert, manager.RouteForDomain("example.com").certificate)
	assert.Equal(t, newCert, manager.RouteForDomain("test.example.com").certificate)
	assert.Equal(t, newCert, manager.RouteForDomain("test.deep.example.com").certificate)
	assert.Equal(t, newCert, manager.RouteForDomain("test.example.net").certificate)
}

func Test_Manager_CheckCertificates_returnsIfNoProvider(t *testing.T) {
	manager := NewManager(nil)
	_ = manager.SetRoutes([]*Route{{
		Domains: []string{"test.deep.example.com", "test.example.com", "example.com"},
	}, {
		Domains: []string{"test.example.net"},
	}})

	err := manager.CheckCertificates()
	assert.NoError(t, err)
}
