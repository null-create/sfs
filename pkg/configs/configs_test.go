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
	return &SvcConfigs{
		configs: cfgsPath,
		env:     make(map[string]string),
		EnvCfgs: env.NewE(),
	}
}

func TestReadYamlFile(t *testing.T) {
	testSvcCfgs := NewTestSvcConfig()

	tcfgs, err := testSvcCfgs.parseYaml()
	if err != nil {
		t.Fail()
	}
	assert.NotEqual(t, nil, tcfgs)
	assert.NotEqual(t, 0, len(tcfgs))
}
