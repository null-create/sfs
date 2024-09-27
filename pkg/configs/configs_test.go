package configs

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sfs/pkg/env"
)

func NewTestSvcConfig() *SvcConfigs {
	cfgsPath, err := filepath.Abs("./testing/test_configs.yaml")
	if err != nil {
		log.Fatal(err)
	}
	cfgs, err := load(cfgsPath)
	if err != nil {
		log.Fatal(err)
	}
	return &SvcConfigs{
		configs: cfgsPath,
		env:     cfgs,
		EnvCfgs: env.NewE(),
	}
}

func TestReadYamlFile(t *testing.T) {
	testSvcCfgs := NewTestSvcConfig()

	assert.NotEqual(t, nil, testSvcCfgs.env)
	assert.NotEqual(t, 0, len(testSvcCfgs.env))
}

func TestUpdateField(t *testing.T) {
	testSvcCfgs := NewTestSvcConfig()

	if err := testSvcCfgs.Set("CLIENT_ID", "1"); err != nil {
		t.Fatal(err)
	}

	cfgs, err := testSvcCfgs.Get("CLIENT_ID")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, 0, len(cfgs))
	assert.Equal(t, "1", cfgs)
}
