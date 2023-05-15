package env

import (
	"os"
)

// PodName example: k8ssta-goapi-main-6868d88fbd-bz8zv
func PodName() string {
	return os.Getenv("PODNAME")
}

// EnvName example: k8ssta
func EnvName() string {
	return os.Getenv("ENV_NAME")
}

// AppName example: worker
func AppName() string {
	return os.Getenv("APP_NAME")
}
