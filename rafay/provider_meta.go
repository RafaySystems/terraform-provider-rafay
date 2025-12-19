package rafay

import (
	rctlconfig "github.com/RafaySystems/rctl/pkg/config"

	v3 "github.com/RafaySystems/rafay-common/pkg/hub/client/typed/infra/v3"
)

type ProviderMeta struct {
	config                 *rctlconfig.Config
	BlueprintClientFactory BlueprintClientFactory
}

type BlueprintClientFactory func() (v3.BlueprintClient, error)

func newProviderMeta(cfg *rctlconfig.Config) *ProviderMeta {
	return &ProviderMeta{
		config: cfg,
	}
}

func (p *ProviderMeta) Config() *rctlconfig.Config {
	if p == nil {
		return nil
	}
	return p.config
}
