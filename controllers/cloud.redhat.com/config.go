package controllers

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"

	clowder "github.com/RedHatInsights/clowder/apis/cloud.redhat.com/v1alpha1"
	frontend "github.com/RedHatInsights/frontend-operator/api/v1alpha1"
	core "k8s.io/api/core/v1"
)

type OperatorConfig struct {
	PoolConfig      PoolConfig                       `yaml:"pool"`
	ClowdEnvSpec    clowder.ClowdEnvironmentSpec     `yaml:"clowdEnv"`
	FrontendEnvSpec frontend.FrontendEnvironmentSpec `yaml:"frontendEnv"`
	LimitRange      core.LimitRange                  `yaml:"limitRange"`
	ResourceQuotas  core.ResourceQuotaList           `yaml:"resourceQuotas"`
}

type PoolConfig struct {
	Size  int  `yaml:"size"`
	Local bool `yaml:"local"`
}

func getConfig() OperatorConfig {
	configPath := "config/environment_config.yaml"
	configMapFile := configPath[strings.LastIndex(configPath, "/")+1:]

	configMapData, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("%s not found\n", configMapFile)
		return OperatorConfig{}
	} else {
		fmt.Printf("Loading config from: %s\n", configMapFile)
	}

	operatorConfig := OperatorConfig{}
	err = yaml.Unmarshal(configMapData, &operatorConfig)
	if err != nil {
		fmt.Printf("Couldn't parse json:\n" + err.Error())
		return OperatorConfig{}
	}

	return operatorConfig
}

var LoadedOperatorConfig OperatorConfig

func init() {
	LoadedOperatorConfig = getConfig()
}
