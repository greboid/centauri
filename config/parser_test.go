package config

import (
	"bytes"
	"testing"

	"github.com/csmith/centauri/proxy"
	"github.com/stretchr/testify/assert"
)

func Test_Parse_ReturnsEmptySliceForEmptyFile(t *testing.T) {
	routes, err := Parse(bytes.NewBuffer([]byte("")))

	assert.NoError(t, err)
	assert.Equal(t, 0, len(routes))
}

func Test_Parse_ErrorsOnUnknownLine(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte("error please")))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnUpstreamOutsideOfRoute(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte("upstream localhost:8080")))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnProviderOutsideOfRoute(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte("provider lego")))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnHeaderOutsideOfRoute(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte("header add x-test foo")))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnHeaderWithTooFewParameters(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte(`
route example.com
	header nothing
`)))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnHeaderAddWithoutValue(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte(`
route example.com
	header add x-test
`)))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnHeaderReplaceWithoutValue(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte(`
route example.com
	header replace x-test
`)))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnHeaderDefaultWithoutValue(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte(`
route example.com
	header default x-test
`)))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnMultipleUpstreams(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte(`
route example.com
	upstream server1
	upstream server2
`)))

	assert.Error(t, err)
}

func Test_Parse_ErrorsOnMultipleProviders(t *testing.T) {
	_, err := Parse(bytes.NewBuffer([]byte(`
route example.com
	provider lego
	provider other
`)))

	assert.Error(t, err)
}

func Test_Parse_ReturnsRoutes(t *testing.T) {
	routes, err := Parse(bytes.NewBuffer([]byte(`
# Comment
route example.com www.example.com
	# Indented comment
	upstream localhost:8080
	header add x-test foo
	header delete x-test-2
	provider p1

route example.net
	upstream localhost:8081
	header default x-test-3 bar
	header replace x-test-4 baz
`)))

	assert.NoError(t, err)
	assert.Equal(t, 2, len(routes))
	assert.Equal(t, []string{"example.com", "www.example.com"}, routes[0].Domains)
	assert.Equal(t, "localhost:8080", routes[0].Upstream)
	assert.Equal(t, "p1", routes[0].Provider)
	assert.Equal(t, []string{"example.net"}, routes[1].Domains)
	assert.Equal(t, "localhost:8081", routes[1].Upstream)
	assert.Equal(t, "", routes[1].Provider)

	// Check headers for the first route
	assert.Equal(t, 2, len(routes[0].Headers))

	assert.Equal(t, "x-test", routes[0].Headers[0].Name)
	assert.Equal(t, "foo", routes[0].Headers[0].Value)
	assert.Equal(t, proxy.HeaderOpAdd, routes[0].Headers[0].Operation)

	assert.Equal(t, "x-test-2", routes[0].Headers[1].Name)
	assert.Equal(t, proxy.HeaderOpDelete, routes[0].Headers[1].Operation)

	// Check headers for the second route
	assert.Equal(t, 2, len(routes[1].Headers))

	assert.Equal(t, "x-test-3", routes[1].Headers[0].Name)
	assert.Equal(t, "bar", routes[1].Headers[0].Value)
	assert.Equal(t, proxy.HeaderOpDefault, routes[1].Headers[0].Operation)

	assert.Equal(t, "x-test-4", routes[1].Headers[1].Name)
	assert.Equal(t, "baz", routes[1].Headers[1].Value)
	assert.Equal(t, proxy.HeaderOpReplace, routes[1].Headers[1].Operation)
}

func Test_Parse_ParsesCaseInsensitively(t *testing.T) {
	routes, err := Parse(bytes.NewBuffer([]byte(`
# Comment
RoUtE example.com www.example.com
	# Indented comment
	UpStReAm localhost:8080
	HeAdEr AdD x-test foo
	hEaDeR dElEtE x-test-2
	PrOvIdEr p1

rOuTe example.net
	uPsTrEaM localhost:8081
	HeAdEr DeFaUlT x-test-3 bar
	hEaDeR rEpLaCe x-test-4 baz
`)))

	assert.NoError(t, err)
	assert.Equal(t, 2, len(routes))
	assert.Equal(t, []string{"example.com", "www.example.com"}, routes[0].Domains)
	assert.Equal(t, "localhost:8080", routes[0].Upstream)
	assert.Equal(t, "p1", routes[0].Provider)
	assert.Equal(t, []string{"example.net"}, routes[1].Domains)
	assert.Equal(t, "localhost:8081", routes[1].Upstream)

	// Check headers for the first route
	assert.Equal(t, 2, len(routes[0].Headers))

	assert.Equal(t, "x-test", routes[0].Headers[0].Name)
	assert.Equal(t, "foo", routes[0].Headers[0].Value)
	assert.Equal(t, proxy.HeaderOpAdd, routes[0].Headers[0].Operation)

	assert.Equal(t, "x-test-2", routes[0].Headers[1].Name)
	assert.Equal(t, proxy.HeaderOpDelete, routes[0].Headers[1].Operation)

	// Check headers for the second route
	assert.Equal(t, 2, len(routes[1].Headers))

	assert.Equal(t, "x-test-3", routes[1].Headers[0].Name)
	assert.Equal(t, "bar", routes[1].Headers[0].Value)
	assert.Equal(t, proxy.HeaderOpDefault, routes[1].Headers[0].Operation)

	assert.Equal(t, "x-test-4", routes[1].Headers[1].Name)
	assert.Equal(t, "baz", routes[1].Headers[1].Value)
	assert.Equal(t, proxy.HeaderOpReplace, routes[1].Headers[1].Operation)
}
