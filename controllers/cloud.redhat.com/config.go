package controllers

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"

	clowder "github.com/RedHatInsights/clowder/apis/cloud.redhat.com/v1alpha1"
	frontend "github.com/RedHatInsights/frontend-operator/api/v1alpha1"
	core "k8s.io/api/core/v1"
)

type ConfigMap struct {
	//	Config OperatorConfig    `yaml:"environment_config.json"`
	Cfg map[string]string `yaml:"data"`
}

type OperatorConfig struct {
	PoolConfig      PoolConfig                       `json:"Pool,omitempty"`
	ClowdEnvSpec    clowder.ClowdEnvironmentSpec     `json:"clowdEnv"`
	FrontendEnvSpec frontend.FrontendEnvironmentSpec `json:"frontendEnv"`
	LimitRange      core.LimitRange                  `json:"limitRange"`
	ResourceQuotas  core.ResourceQuotaList           `json:"resourceQuotas"`
}

type PoolConfig struct {
	Size  int  `json:"size"`
	Local bool `json:"local"`
}

func getConfig() OperatorConfig {
	configMapPath := "config/environment_config.yaml"
	configMapFileName := configMapPath[strings.LastIndex(configMapPath, "/")+1:]

	configMapData, err := ioutil.ReadFile(configMapPath)
	if err != nil {
		fmt.Printf("%s not found\n", configMapFileName)
		return OperatorConfig{}
	} else {
		fmt.Printf("Loading config from: %s\n", configMapFileName)
	}

	configMap := ConfigMap{}
	operatorConfig := OperatorConfig{}
	err = yaml.Unmarshal(configMapData, &configMap)
	if err != nil {
		fmt.Printf("Couldn't parse json:\n" + err.Error())
		return OperatorConfig{}
	}

	err = json.Unmarshal([]byte(configMap.Cfg["environment_config.json"]), &operatorConfig)

	return operatorConfig
}

var LoadedOperatorConfig OperatorConfig

func init() {
	LoadedOperatorConfig = getConfig()
}
