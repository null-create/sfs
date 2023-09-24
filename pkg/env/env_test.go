package env

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func buildTestEnvMap() map[string]string {
	env := make(map[string]string)
	env["NEW_SERVICE"] = "true"
	env["ADMIN_MODE"] = "false"
	env["JWT_SECRET"] = "secret-goes-here"
	env["SERVICE_ROOT"] = "C:\\Users\\Jay\\Google Drive\\Projects\\coding\\nimbus\\pkg\\server"
	env["SERVER_PORT"] = "8080"
	env["SERVER_HOST"] = "localhost"
	env["SERVER_ADDR"] = "localhost:8080"
	env["SERVER_TIMEOUT_READ"] = "5s"
	env["SERVER_TIMEOUT_WRITE"] = "10s"
	env["SERVER_TIMEOUT_IDLE"] = "900s"
	env["SERVER_ADMIN"] = "admin"
	env["SERVER_ADMIN_KEY"] = "derp"
	return env
}

func TestEnvBuild(t *testing.T) {
	BuildEnv()
	testEnv := buildTestEnvMap()
	for k, env := range testEnv {
		v := os.Getenv(k)
		assert.Equal(t, v, env)
	}
}

func TestEnvValidate(t *testing.T) {
	BuildEnv()
	e := NewE()
	if err := e.Validate("JWT_SECRET"); err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, nil, e)
}
