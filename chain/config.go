package chain

// AssetsConfig 链配置加载接口（可选），由 ChainAdapter 聚合。
type AssetsConfig interface {
	LoadAssetsConfig(config interface{}) error
	InitAssetsConfig() (interface{}, error)
}

// AssetsConfigBase 默认空实现
type AssetsConfigBase struct{}

func (AssetsConfigBase) LoadAssetsConfig(config interface{}) error { return nil }
func (AssetsConfigBase) InitAssetsConfig() (interface{}, error)    { return nil, nil }
