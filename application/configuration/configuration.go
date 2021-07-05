package configuration

var appConfig *Config

func SetAppConfig(config Config) {
	if !appConfig.set {
		appConfig = &config
		appConfig.set = true
	}
}

func GetAppConfig() *Config {
	return appConfig
}
