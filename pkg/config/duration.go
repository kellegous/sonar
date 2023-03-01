package config

import (
	"time"

	"github.com/BurntSushi/toml"
)

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	var v string
	if err := toml.Unmarshal(text, &v); err != nil {
		return err
	}
	t, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*d = Duration(t)
	return nil
}
