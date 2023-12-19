package config

func init() {
	Define("config.path")

	Flag.Define("config.readonly", false)
	Flag.Define("config.wal", true)
}
