package xray

import (
	"encoding/json"
	"guardhelper/internal/config"
	"io"
	"log"
	"os"
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
	for _, inbound := range xrayConfig.Inbounds {
		if inbound.Tag == "" || inbound.Protocol == "" {
			continue
		}
		inboundsByProtocol[inbound.Protocol] = append(inboundsByProtocol[inbound.Protocol], inbound.Tag)
	}

	return inboundsByProtocol, nil
}
