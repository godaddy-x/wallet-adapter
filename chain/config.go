package chain

// AssetsConfig chain config loading interface (optional), aggregated by ChainAdapter.
type AssetsConfig interface {
	LoadAssetsConfig(config interface{}) error
	InitAssetsConfig() (interface{}, error)
}

// AssetsConfigBase default empty implementation.
type AssetsConfigBase struct{}

func (AssetsConfigBase) LoadAssetsConfig(config interface{}) error { return nil }
func (AssetsConfigBase) InitAssetsConfig() (interface{}, error)    { return nil, nil }
