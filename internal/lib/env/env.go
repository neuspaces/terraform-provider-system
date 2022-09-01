package env

import "os"

// Has returns true is the provided environment variable is set
func Has(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func Set(key string, value string) {
	if err := os.Setenv(key, value); err != nil {
		panic(err)
	}
}

// SetOptional sets the value of the environment variable named by the key if the environment variable is not set.
// SetOptional panics if os.Setenv fails with an error.
func SetOptional(key string, value string) {
	if _, ok := os.LookupEnv(key); !ok {
		if err := os.Setenv(key, value); err != nil {
			panic(err)
		}
	}
}
