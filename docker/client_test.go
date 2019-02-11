package docker

import (
	"os"
	"reflect"
	"testing"

	"github.com/docker/docker/client"
)

const (
	endpoint   = "unix:///var/run/docker.sock"
	certPath   = "testdata/cert.pem"
	keyPath    = "testdata/key.pem"
	caPath     = "testdata/ca.pem"
	apiVersion = "1.39"
)

func TestNewClientConfig(t *testing.T) {
	config := &ClientConfig{
		endpoint:   endpoint,
		certPath:   certPath,
		keyPath:    keyPath,
		caPath:     caPath,
		apiVersion: apiVersion,
	}

	verify := NewClientConfig(endpoint, certPath, keyPath, caPath, apiVersion)
	if config.endpoint != verify.endpoint {
		t.Errorf("Expected endpoint %+v. Got %+v.", config.endpoint, verify.endpoint)
	}
	if config.certPath != verify.certPath {
		t.Errorf("Expected certPath %+v. Got %+v.", config.certPath, verify.certPath)
	}
	if config.keyPath != verify.keyPath {
		t.Errorf("Expected keyPath %+v. Got %+v.", config.keyPath, verify.keyPath)
	}
	if config.caPath != verify.caPath {
		t.Errorf("Expected caPath %+v. Got %+v.", config.caPath, verify.caPath)
	}
}

func TestNewDocker(t *testing.T) {
	client, err := client.NewClientWithOpts(client.WithHost(endpoint), client.WithVersion(apiVersion))
	if err != nil {
		t.Fatal(err)
	}
	dockerClient := &Docker{client}

	config := NewClientConfig(endpoint, "", "", "", "")
	verify := NewDocker(config)

	expect := reflect.ValueOf(dockerClient).Elem().FieldByName("endpoint").String()
	got := reflect.ValueOf(verify).Elem().FieldByName("endpoint").String()

	if expect != got {
		t.Errorf("Expected endpoint %s. Got %s.", expect, got)
	}
}

func TestNewDockerTLS(t *testing.T) {
	client, err := client.NewClientWithOpts(client.WithTLSClientConfig(caPath, certPath, keyPath), client.WithHost(endpoint), client.WithVersion(apiVersion))
	if err != nil {
		t.Fatal(err)
	}
	dockerClient := &Docker{client}

	config := NewClientConfig(endpoint, certPath, keyPath, caPath, apiVersion)
	verify := NewDocker(config)

	expect := reflect.ValueOf(dockerClient).Elem().FieldByName("endpoint").String()
	got := reflect.ValueOf(verify).Elem().FieldByName("endpoint").String()

	if expect != got {
		t.Errorf("Expected endpoint %s. Got %s.", expect, got)
	}
}

func TestNewDockerFromEnv(t *testing.T) {
	os.Setenv("DOCKER_HOST", endpoint)
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(apiVersion))
	if err != nil {
		t.Fatal(err)
	}

	dockerClient := &Docker{client}
	config := NewClientConfig("dummy endpoint", "", "", "", "")
	verify := NewDocker(config)

	if reflect.DeepEqual(dockerClient, verify) {
		t.Errorf("Expected Docker env clinet %+v. Got %+v.", dockerClient.Client, verify.Client)
	}
}
