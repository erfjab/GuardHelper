package xray

import (
	"encoding/json"
	"guardhelper/internal/config"
	"io"
	"log"
	"os"
	"strings"
)

type XrayConfig struct {
	Inbounds []Inbound `json:"inbounds"`
}

type Inbound struct {
	Tag      string `json:"tag"`
	Protocol string `json:"protocol"`
}

func GetInboundsByProtocol() (map[string][]string, error) {
	configPath := config.Cfg.XrayConfigPath

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Xray config file not found at %s", configPath)
		return map[string][]string{}, err 
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var xrayConfig XrayConfig
	if err := json.Unmarshal(bytes, &xrayConfig); err != nil {
		log.Printf("Error unmarshalling xray config: %v", err)
		return nil, err
	}

	inboundsByProtocol := make(map[string][]string)
	seen := make(map[string]map[string]bool)
	for _, inbound := range xrayConfig.Inbounds {
		protocol := strings.ToLower(strings.TrimSpace(inbound.Protocol))
		tag := strings.TrimSpace(inbound.Tag)
		if tag == "" || protocol == "" {
			continue
		}

		if seen[protocol] == nil {
			seen[protocol] = make(map[string]bool)
		}
		if seen[protocol][tag] {
			continue
		}
		seen[protocol][tag] = true
		inboundsByProtocol[protocol] = append(inboundsByProtocol[protocol], tag)
	}

	return inboundsByProtocol, nil
}
