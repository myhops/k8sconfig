package k8sconfig

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

const schemeName = "k8scfg"

type provider struct {
	// Add k8s credentials
	logger *zap.Logger
}

// valueSpec contains the result of the parseURI
type valueSpec struct {
	kind string
	namespace string
	name string
	dataType string
	key string
}

func NewFactory() confmap.ProviderFactory {
	return confmap.NewProviderFactory(newProvider)
}

func newProvider(settings confmap.ProviderSettings) confmap.Provider {
	return &provider{
		logger: settings.Logger,
	}
}

func parseURI(uri string) (*valueSpec, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// Test scheme
	if u.Scheme != schemeName {
		return nil, fmt.Errorf("expected k8scfg scheme")
	}
	// Opaque should be set
	if u.Opaque == "" {
		return nil, fmt.Errorf("not opaque")
	}

	parts := strings.Split(u.Opaque, "/")
	if len(parts) != 5 {
		return nil, fmt.Errorf("need 5 parts, got %d", len(u.Opaque))
	}

	return &valueSpec{
		kind: parts[0],
		namespace: parts[1],
		name: parts[2],
		dataType: parts[3],
		key: parts[4],
	}, nil
}

func (v *valueSpec) validate() error {
	return nil
}

func getFromConfigMap(ctx context.Context, namespace string, name string, field string, item string) ([]byte, error) {
	// first check the field type
	if field != "data" && field != "binaryData" {
		return nil, fmt.Errorf("field must be either data or binaryData")
	}

	cm, err := getConfigMap(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	var res []byte
	switch field {
	case "data":
		res = []byte(cm.Data[item])
	case "binaryData":
		res = cm.BinaryData[item]
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("%s[%s] is empty", field, item)
	}
	return res, nil
}

func getFromSecret(ctx context.Context, namespace string, name string, field string, item string) ([]byte, error) {
	// first check the field type
	if field != "data" && field != "binaryData" {
		return nil, fmt.Errorf("field must be either data or binaryData")
	}

	s, err := getSecret(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	var res []byte
	switch field {
	case "data":
		res = s.Data[item]
	case "binaryData":
		res = []byte(s.StringData[item])
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("%s[%s] is empty", field, item)
	}
	return res, nil
}

// Retreive gets the data from either a secret or a config map using the following schema
//
//	k8sconfig:configMap:namespace:<configmap name>:<data or binaryData>:<item>  for a configmap
//	k8sconfig:secret:namespace:<secret name>:<data or stringData>:<item>  for a secret
func (p *provider) Retrieve(ctx context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {

	// use uri to parse

	parts := strings.Split(uri, ":")
	if l := len(parts); l != 6 {
		return nil, fmt.Errorf("number of parts must be 6, got %v", l)
	}

	if parts[0] != schemeName {
		return nil, fmt.Errorf("%q uri is not supported by %q provider", uri, schemeName)
	}

	var res []byte
	var err error
	switch strings.ToLower(parts[1]) {
	case "configmap":
		res, err = getFromConfigMap(ctx, parts[2], parts[3], parts[4], parts[5])
	case "secret":
		res, err = getFromSecret(ctx, parts[2], parts[3], parts[4], parts[5])
	}
	if err != nil {
		return nil, err
	}

	return confmap.NewRetrievedFromYAML(res)
}

func (*provider) Scheme() string {
	return schemeName
}

func (*provider) Shutdown(context.Context) error {
	return nil
}
