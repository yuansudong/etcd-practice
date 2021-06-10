package center

type System struct {
	Mysql string `yaml:"mysql"`
	Redis string `yaml:"redis"`
	Nats  string `yaml:"nats"`
	Nsqd  string `yaml:"nsqd"`
}
