package log

import "github.com/rs/zerolog"

type Config struct {
	BizName string        `yaml:"biz_name"`
	File    string        `yaml:"file"`
	Level   zerolog.Level `yaml:"level"`
}
