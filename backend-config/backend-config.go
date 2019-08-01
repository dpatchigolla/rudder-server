package backendconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-server/config"
	"github.com/rudderlabs/rudder-server/utils"
)

var (
	configBackendURL, configBackendToken string
	pollInterval                         time.Duration
)

var Eb *utils.EventBus

type DestinationDefinitionT struct {
	ID   string
	Name string
}

type SourceDefinitionT struct {
	ID   string
	Name string
}

type DestinationT struct {
	ID                    string
	Name                  string
	DestinationDefinition DestinationDefinitionT
	Config                interface{}
	Enabled               bool
}

type SourceT struct {
	ID               string
	Name             string
	SourceDefinition SourceDefinitionT
	Config           interface{}
	Enabled          bool
	Destinations     []DestinationT
	WriteKey         string
}

type SourcesT struct {
	Sources []SourceT `json:"sources"`
}

func loadConfig() {
	configBackendURL = config.GetEnv("CONFIG_BACKEND_URL", "http://localhost:3000")
	configBackendToken = config.GetEnv("CONFIG_BACKEND_TOKEN", "1OjNiOZZhLc0EcaQv2IhpA2IwoC")
	pollInterval = config.GetDuration("BackendConfig.pollIntervalInS", 5) * time.Second
}

func getBackendConfig() SourcesT {
	client := &http.Client{}
	url := fmt.Sprintf("%s/workspace-config?workspaceToken=%s", configBackendURL, configBackendToken)
	resp, err := client.Get(url)

	var respBody []byte
	if resp != nil && resp.Body != nil {
		respBody, _ = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
	}

	if err != nil {
		log.Println("Errored when sending request to the server", err)
	}
	var sourcesJSON SourcesT
	err = json.Unmarshal(respBody, &sourcesJSON)
	return sourcesJSON
}

func init() {
	config.Initialize()
	loadConfig()
}

func pollConfigUpdate() {
	for {
		sourceJSON := getBackendConfig()
		Eb.Publish("backendconfig", sourceJSON)
		time.Sleep(time.Duration(pollInterval))
	}
}

// Setup backend config
func Setup() {
	Eb = new(utils.EventBus)
	go pollConfigUpdate()
}