package client

import (
	"log"

	"github.com/sfs/pkg/env"

	"github.com/joeshaw/envdecode"
)

var EndpointRoot = "http://" + cfgs.Host

type Conf struct {
	IsAdmin         bool   `env:"ADMIN_MODE"`                   // whether the service should be run in admin mode or not
	BufferedEvents  bool   `env:"BUFFERED_EVENTS,required"`     // whether events should be buffered (i.e. have a delay between sync events)
	EventBufferSize int    `env:"EVENT_BUFFER_SIZE,required"`   // size of events buffer
	User            string `env:"CLIENT_NAME,required"`         // users name
	UserAlias       string `env:"CLIENT_USERNAME,required"`     // users alias (username)
	ID              string `env:"CLIENT_ID,required"`           // this is generated at creation time. won't be in the initial .env file
	Email           string `env:"CLIENT_EMAIL,required"`        // users email
	Password        string `env:"CLIENT_PASSWORD,required"`     // users password for authentication
	Root            string `env:"CLIENT_ROOT,required"`         // client service root (ie. ../sfs/client/run/)
	TestRoot        string `env:"CLIENT_TESTING,required"`      // testing root directory
	Host            string `env:"CLIENT_HOST,required"`         // client host
	Port            int    `env:"CLIENT_PORT,required"`         // port for http client
	Addr            string `env:"CLIENT_ADDRESS,required"`      // address for http client
	NewService      bool   `env:"CLIENT_NEW_SERVICE,required"`  // whether we need to initialize a new client service instance.
	LogDir          string `env:"CLIENT_LOG_DIR,required"`      // location of log directory
	LocalBackup     bool   `env:"CLIENT_LOCAL_BACKUP,required"` // whether we're backing up to a local file directory
	BackupDir       string `env:"CLIENT_BACKUP_DIR,required"`   // location of backup directory
}

func GetClientConfigs() *Conf {
	env.SetEnv(false)

	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("failed to decode client config .env file: %s", err)
	}
	return &c
}

// client env, user, and service configurations
var cfgs = GetClientConfigs()

// Environment configs
var envCfgs = env.NewE()
