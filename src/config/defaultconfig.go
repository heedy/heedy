package config

import (
	"encoding/base64"

	"github.com/dkumor/acmewrapper"
	"github.com/gorilla/securecookie"
	"github.com/nu7hatch/gouuid"

	psconfig "github.com/connectordb/pipescript/config"
)

// NewConfiguration generates a configuration with reasonable defaults for use in ConnectorDB
func NewConfiguration() *Configuration {
	redispassword, _ := uuid.NewV4()
	natspassword, _ := uuid.NewV4()

	sessionAuthKey := securecookie.GenerateRandomKey(64)
	sessionEncKey := securecookie.GenerateRandomKey(32)

	return &Configuration{
		Version:     1,
		Watch:       true,
		Permissions: "default",
		Redis: Service{
			Hostname: "localhost",
			Port:     6380,
			Password: redispassword.String(),
			Enabled:  true,
		},
		Nats: Service{
			Hostname: "localhost",
			Port:     4223,
			Username: "connectordb",
			Password: natspassword.String(),
			Enabled:  true,
		},
		Sql: &SQLService{
			Type: "postgres",
			URI:  "", // connectordb generates a postgres uri if not given
			Service: Service{
				Hostname: "localhost",
				Port:     52593,
				//TODO: Have SQL access be auth'd
				Enabled: true,
			},
		},

		Frontend: Frontend{
			Hostname: "",   // Host on all interfaces by default
			Port:     8000, // Port 8000 by default

			Redirect80: false,

			Enabled: true,

			LogFile:  "",
			LogLevel: "info",

			// Sets up the session cookie keys that are used
			CookieSession: CookieSession{
				AuthKey:       base64.StdEncoding.EncodeToString(sessionAuthKey),
				EncryptionKey: base64.StdEncoding.EncodeToString(sessionEncKey),
				MaxAge:        60 * 60 * 24 * 30 * 4, //About 4 months is the default expiration time of a cookie
				Remember:      true,
			},

			// By default, captcha is disabled
			Captcha: Captcha{
				Enabled: false,
			},

			// Set up the default TLS options
			TLS: TLS{
				Enabled: false,
				Key:     "tls_key.key",
				Cert:    "tls_cert.crt",
				ACME: ACME{
					Server:       acmewrapper.DefaultServer,
					PrivateKey:   "acme_privatekey.pem",
					Registration: "acme_registration.json",
					Domains:      []string{"example.com", "www.example.com"},
					Enabled:      false,
				},
			},

			// By default log query counts once a minute, and display server statistics
			// once a day
			QueryDisplayTimer: 60,
			StatsDisplayTimer: 60 * 60 * 24,

			// A limit of 10MB of data per insert is reasonable to me
			InsertLimitBytes: 1024 * 1024 * 10,

			// The options that pertain to the websocket interface
			Websocket: Websocket{
				// 1MB per websocket is also reasonable
				MessageLimitBytes: 1024 * 1024,

				// The time to wait on a socket write in seconds
				WriteWait: 2,

				// Websockets ping each other to keep the connection alive
				// This sets the number of seconds between pings
				PongWait:   60,
				PingPeriod: 54,

				// Websocket upgrader read/write buffer sizes
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,

				// 3 messages should be enough... right?
				MessageBuffer: 3,
			},

			// Why not minify? Turning it off is useful for debugging - but users outnumber coders by a large margin.
			Minify: true,

			// These files don't have any sensitive data, so it should be OK to send them compressed over https
			CacheStatic:    true,
			CacheStaticAge: 604800,
			GzipStatic:     true,

			//wait a full second between authentication failures
			FailedLoginDelay: 300,
		},

		//The defaults to use for the batch and chunks
		BatchSize: 250,
		ChunkSize: 5,

		UseCache:        true,
		UserCacheSize:   1000,
		DeviceCacheSize: 10000,
		StreamCacheSize: 10000,

		// This is the CONSTANT default. The database will explode if this is ever changed.
		// You have been warned.
		IDScramblePrime: 2147483423,

		// No reason not to use bcrypt
		PasswordHash: "bcrypt",

		// Use the default settings.
		PipeScript: psconfig.Default(),
	}

}
