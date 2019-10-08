package authrlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testCfg = `
app:
  name: "Authorization Service"
  env: "dev"
  port: 8888
  id: "1"
  keys-server: "http://localhost:8888"
  newrelic:
  enabled: false
  apikey: "1234567890123456789012345678901234567890"
  ignorehttpcodes: [400, 401, 402, 403, 404, 405, 406]
  mssql:
    connection:
      host: "10.0.0.53"
      port: 62433
      database: "CCNET"
      user: "test user"
      password: "mysecret"`

func TestConfigService(t *testing.T) {
	// error
	configFile := "/tmp/fake/path/to/config/file.yaml"
	assert.Panics(t, func() { NewApplicationConfigService(&configFile) })

	// success
	file, err := ioutil.TempFile(os.TempDir(), "cfg")
	if err != nil {
		t.Fatal("unable to create tmp file for test", err)
	}
	configFile = file.Name()
	err = ioutil.WriteFile(configFile, []byte(testCfg), 0644)
	if err != nil {
		t.Fatal("unable to write test data into file", err)
	}
	configService := NewApplicationConfigService(&configFile)
	assert.NotNil(t, configService)
	assert.NotNil(t, configService.Config())
	assert.False(t, configService.IsProduction())
	assert.Equal(t, "8888", configService.Config().App.PortToStr())
}

func TestInitialize(t *testing.T) {
	config := &Config{}

	// file not found
	configFile := "/tmp/fake/path/to/config/file.yaml"
	err := initialize(&configFile, config)
	expected := "unable to find application configuration file: " + configFile
	if err.Error() != expected {
		t.Fatalf("Got '%s', expected '%s'", err.Error(), expected)
	}

	// unable to read file
	dir, err := ioutil.TempDir(os.TempDir(), "prefix")
	if err != nil {
		t.Fatal("unable to create tmp dir for test", err)
	}
	configFile = dir
	err = initialize(&configFile, config)
	expected = fmt.Sprintf("read %s: is a directory", dir)
	if expected != err.Error() {
		t.Fatalf("Got '%s', expected '%s'", err.Error(), expected)
	}

	// unable to unmarshal
	file, err := ioutil.TempFile(dir, "cfg")
	if err != nil {
		t.Fatal("unable to create file for test", err)
	}
	configFile = file.Name()
	err = ioutil.WriteFile(configFile, []byte(`Fake\n`), 0644)
	if err != nil {
		t.Fatal("unable to write test data into file", err)
	}
	err = initialize(&configFile, config)
	expected = "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `Fake\\n` into authrlib.Config"
	if expected != err.Error() {
		t.Fatalf("Got '%s', expected '%s'", err.Error(), expected)
	}
}
