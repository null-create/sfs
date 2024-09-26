package configs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sfs/pkg/env"

	"gopkg.in/yaml.v3"
)

/*
This package implements the following interfaces so that
we can replace envCfgs in the client module with this package, rather
than using the env module itself

func (e *Env) Clear() error
func (e *Env) Get(k string) (string, error)
func (e *Env) GetEnv() (map[string]string, error)
func (e *Env) List()
func (e *Env) Set(k string, v string) error
func (e *Env) Validate(k string) error
*/

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
	return &SvcConfigs{
		configs: cfgsPath,
		env:     make(map[string]string),
		EnvCfgs: env.NewE(),
	}
}

var svcCfgs = NewSvcConfig()

func (s *SvcConfigs) load() (map[string]string, error) {
	cfgsFile, err := os.Open(s.configs)
	if err != nil {
		return nil, err
	}
	defer cfgsFile.Close()

	cfgs := make(map[string]string)
	if err := yaml.NewDecoder(cfgsFile).Decode(&cfgs); err != nil {
		return nil, err
	}
	s.env = cfgs
	return cfgs, nil
}

// used for setting environment configurations at runtime
func SetEnv(debug bool) error {
	if _, err := svcCfgs.load(); err != nil {
		return err
	}
	env.SetEnv(debug)
	return nil
}

// EnvCfgs wrappers

func (s *SvcConfigs) Clear() error                       { return s.EnvCfgs.Clear() }
func (s *SvcConfigs) Get(k string) (string, error)       { return s.EnvCfgs.Get(k) }
func (s *SvcConfigs) GetEnv() (map[string]string, error) { return s.EnvCfgs.GetEnv() }
func (s *SvcConfigs) List()                              { s.EnvCfgs.List() }
func (s *SvcConfigs) Validate(k string) error            { return s.EnvCfgs.Validate(k) }

func (s *SvcConfigs) Set(k string, v string) error {
	if _, ok := s.env[k]; ok {
		if err := s.EnvCfgs.Set(k, v); err != nil {
			return err
		}
		cfgsFile, err := os.Open(s.configs)
		if err != nil {
			return err
		}
		defer cfgsFile.Close()

		s.env[k] = v
		return yaml.NewEncoder(cfgsFile).Encode(s.env)
	} else {
		return fmt.Errorf("key not found: %v", k)
	}
}
