package rafay

import (
	rctlconfig "github.com/RafaySystems/rctl/pkg/config"

	v3 "github.com/RafaySystems/rafay-common/pkg/hub/client/typed/infra/v3"
)

type providerMeta struct {
	config                 *rctlconfig.Config
	blueprintClientFactory blueprintClientFactory
}

type blueprintClientFactory func() (v3.BlueprintClient, error)

func newProviderMeta(cfg *rctlconfig.Config) *providerMeta {
	return &providerMeta{
		config: cfg,
	}
}

func (p *providerMeta) Config() *rctlconfig.Config {
	if p == nil {
		return nil
	}
	return p.config
}
