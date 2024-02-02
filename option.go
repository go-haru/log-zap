package logger

type Options struct {
	Level     string `json:"level" yaml:"level"`
	Format    string `json:"format" yaml:"format"`
	LongTime  bool   `json:"longTime" yaml:"longTime"`
	WithColor bool   `json:"withColor" yaml:"withColor"`
}
