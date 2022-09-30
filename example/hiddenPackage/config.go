package hiddenPackage

type Config struct {
	_LOCKED bool
	uri     string
}

func (m *Config) Uri() string {
	return m.uri
}
