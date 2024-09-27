package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/env"

	"gopkg.in/yaml.v3"
)

type SvcConfigs struct {
	configs string            // path to the yaml config file
	env     map[string]string // parsed environment variable from yaml config file
	EnvCfgs *env.Env          // environment variables
}

func NewSvcConfig() *SvcConfigs {
	cfgsPath, err := filepath.Abs("./pkg/configs/configs.yaml")
	if err != nil {
		log.Fatal(err)
	}
	envCfgs, err := load(cfgsPath)
	if err != nil {
		log.Fatal(err)
	}
	return &SvcConfigs{
		configs: cfgsPath,
		env:     envCfgs,
		EnvCfgs: env.NewE(),
	}
}

// typically ran at project root level during runtime
func getroot() string {
	path, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	return path
}

func load(cfgsFile string) (map[string]string, error) {
	file, err := os.ReadFile(cfgsFile)
	if err != nil {
		return nil, err
	}

	cfgs := make(map[string]string)
	if err := yaml.Unmarshal(file, &cfgs); err != nil {
		return nil, err
	}
	return cfgs, nil
}

// used for setting environment configurations at runtime
func SetEnv(debug bool) error {
	if !env.HasEnvFile(getroot()) {
		fmt.Printf("missing environment files. run 'sfs setup' first.")
		os.Exit(1)
	}
	env.SetEnv(debug)
	return nil
}

func (s *SvcConfigs) Set(k, v string) error {
	if _, ok := s.env[k]; !ok {
		return fmt.Errorf("key '%s' not found", k)
	}
	if err := s.EnvCfgs.Set(k, v); err != nil {
		return err
	}
	cfgsFile, err := os.OpenFile(s.configs, os.O_RDWR|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer cfgsFile.Close()

	s.env[k] = v
	return yaml.NewEncoder(cfgsFile).Encode(&s.env)
}

// EnvCfgs wrappers
func (s *SvcConfigs) Clear() error                 { return s.EnvCfgs.Clear() }
func (s *SvcConfigs) Get(k string) (string, error) { return s.EnvCfgs.Get(k) }
func (s *SvcConfigs) List()                        { s.EnvCfgs.List() }
