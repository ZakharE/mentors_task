package config

import "sync"

type config struct {
	AccessTokenExpiresIn  int64
	RefreshTokenExpiresIn int64
}

var once= sync.Once{}

var cfg *config

func Get() *config {
	once.Do(func() {
		cfg= &config{}
	})
	return cfg
}
