package chain

// VerifyAPIsConfigurer chains that support Phase C peer pools (verifyAPIs in nodeConfig).
type VerifyAPIsConfigurer interface {
	SetVerifyAPIs(urls []string)
}
