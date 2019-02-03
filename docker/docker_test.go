package docker

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	endpoint = "unix:///var/run/docker.sock"
	certPath = "testdata/cert.pem"
	keyPath  = "testdata/key.pem"
	caPath   = "testdata/key.pem"
)

func TestNewClientConfig(t *testing.T) {
	config := &ClientConfig{
		endpoint: endpoint,
		certPath: certPath,
		keyPath:  keyPath,
		caPath:   caPath,
	}

	verify := NewClientConfig(endpoint, certPath, keyPath, caPath)
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
	client, err := docker.NewClient(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	dockerClient := &Docker{client}

	config := NewClientConfig(endpoint, "", "", "")
	verify := NewDocker(config)

	if reflect.DeepEqual(dockerClient, verify) {
		t.Errorf("Expected Docker clinet %+v. Got %+v.", dockerClient.Client, verify.Client)
	}
}

func TestNewDockerTLS(t *testing.T) {
	client, err := docker.NewTLSClient(endpoint, certPath, keyPath, caPath)
	if err != nil {
		t.Fatal(err)
	}
	dockerClient := &Docker{client}

	config := NewClientConfig(endpoint, certPath, keyPath, caPath)
	verify := NewDocker(config)

	if reflect.DeepEqual(dockerClient, verify) {
		t.Errorf("Expected Docker TLS clinet %+v. Got %+v.", dockerClient.Client, verify.Client)
	}
}

func TestNewDockerFromEnv(t *testing.T) {
	os.Setenv("DOCKER_HOST", endpoint)

	client, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	dockerClient := &Docker{client}
	config := NewClientConfig("dummy endpoint", "", "", "")
	verify := NewDocker(config)

	if reflect.DeepEqual(dockerClient, verify) {
		t.Errorf("Expected Docker env clinet %+v. Got %+v.", dockerClient.Client, verify.Client)
	}
}

func TestNewDockerFromEnvTLS(t *testing.T) {
	base, _ := os.Getwd()
	os.Setenv("DOCKER_CERT_PATH", filepath.Join(base, "/testdata/"))
	os.Setenv("DOCKER_HOST", endpoint)
	os.Setenv("DOCKER_TLS_VERIFY", "1")

	client, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	dockerClient := &Docker{client}
	config := NewClientConfig("dummy endpoint", "dummy cert", "dummy key", "dummy ca")
	verify := NewDocker(config)

	if reflect.DeepEqual(dockerClient, verify) {
		t.Errorf("Expected Docker env TLS clinet %+v. Got %+v.", dockerClient.Client, verify.Client)
	}
}
