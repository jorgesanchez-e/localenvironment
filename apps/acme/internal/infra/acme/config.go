package acme

type implementedDriver int

const (
	awsDriver implementedDriver = iota
)

func (imp implementedDriver) String() string {
	return [...]string{"AWS"}[imp]
}

type Configurator interface {
	Get(key string, value any) error
}

type acmeConfig struct {
	LogLeve    string        `mapstructure:"log-level"`
	Email      string        `mapstructure:"email"`
	PrimaryKey string        `mapstructure:"primary_key"`
	AWSBackend []*awsBackEnd `mapstructure:"aws"`
}

type awsBackEnd struct {
	AWS any
}

func (ac acmeConfig) implementations() []implementedDriver {
	implementations := make([]implementedDriver, 0)

	if ac.AWSBackend != nil {
		implementations = append(implementations, awsDriver)
	}

	return implementations
}
