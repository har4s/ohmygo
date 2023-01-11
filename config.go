package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Env struct {
	IsDebug        bool
	Port           int // default 8000
	DatabaseURL    string
	SkipMigrations bool // default false
}

func NewEnv() *Env {
	_, isUsingGoRun := inspectRuntime()

	env := &Env{
		IsDebug:        isUsingGoRun,
		Port:           8000,
		DatabaseURL:    "root:@/example",
		SkipMigrations: false,
	}

	if v, ok := env.GetBool("DEBUG"); ok {
		env.IsDebug = v
	}

	if v, ok := env.GetInt("PORT"); ok {
		env.Port = v
	}

	if v, ok := env.Get("DATABASE_URL"); ok {
		env.DatabaseURL = v
	}

	if v, ok := env.GetBool("SKIP_MIGRATIONS"); ok {
		env.SkipMigrations = v
	}

	return env
}

func (env *Env) Get(key string) (string, bool) {
	val := os.Getenv(key)
	if val == "" {
		return "", false
	}
	return val, true
}

func (env *Env) GetBool(key string) (bool, bool) {
	val, ok := env.Get(key)
	if !ok {
		return false, false
	}
	return (strings.ToLower(val) == "true" || val == "1"), true
}

func (env *Env) GetInt(key string) (int, bool) {
	val, ok := env.Get(key)
	if !ok {
		return 0, false
	}
	valInt, err := strconv.Atoi(val)
	return valInt, err == nil
}

// inspectRuntime tries to find the base executable directory and how it was run.
func inspectRuntime() (baseDir string, withGoRun bool) {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// probably ran with go run
		withGoRun = true
		baseDir, _ = os.Getwd()
	} else {
		// probably ran with go build
		withGoRun = false
		baseDir = filepath.Dir(os.Args[0])
	}
	return
}
