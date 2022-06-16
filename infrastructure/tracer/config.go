package tracer

type Config struct {
	File      string  `yaml:"file"`
	ReportUrl string  `yaml:"report_url"`
	Rate      float64 `yaml:"rate"`
}
