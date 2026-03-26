package factory

// AdrfConfig keeps the process-wide loaded configuration object.
//
// This matches free5gc-style global config access and allows packages that do
// not own initialization flow (e.g. SBI helpers) to access static settings.
var AdrfConfig *Config
