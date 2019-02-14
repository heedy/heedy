package config

import (
	"crypto/tls"
	"errors"
	"path/filepath"
)

// ACME is the struct which holds the Let's Encrypt options of the database
type ACME struct {
	Enabled      bool     `json:"enabled"`
	Server       string   `json:"server"`
	PrivateKey   string   `json:"private_key"`
	Registration string   `json:"registration"`
	Domains      []string `json:"domains"`
	TOSAgree     bool     `json:"tos_agree"`
}

// TLS enables TLS support on the server
type TLS struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
	Cert    string `json:"cert"`

	ACME ACME `json:"acme"`
}

// Validate ensures that the TLS configuration is OK
func (t *TLS) Validate() (err error) {
	if t.Enabled {
		if t.Key == "" || t.Cert == "" {
			return errors.New("TLS key or cert was not given")
		}
		//Set the file paths to be full paths
		t.Cert, err = filepath.Abs(t.Cert)
		if err != nil {
			return err
		}
		t.Key, err = filepath.Abs(t.Key)
		if err != nil {
			return err
		}

		if t.ACME.Enabled {

			if t.ACME.PrivateKey == "" || t.ACME.Registration == "" {
				return errors.New("ACME registration and private key files not given")
			}

			t.ACME.PrivateKey, err = filepath.Abs(t.ACME.PrivateKey)
			if err != nil {
				return err
			}
			t.ACME.Registration, err = filepath.Abs(t.ACME.Registration)
			if err != nil {
				return err
			}

			if t.ACME.Server == "" {
				return errors.New("ACME server not given")
			}
			if len(t.ACME.Domains) == 0 {
				return errors.New("ACME requires a valid list of domains for certificate")
			}
			if !t.ACME.TOSAgree {
				return errors.New("Must agree to the TOS of your ACME server.")
			}

		} else {
			// If ACME is not on, we require that the key/cert exist already
			_, err = tls.LoadX509KeyPair(t.Cert, t.Key)
			if err != nil {
				return err
			}
		}

		//

	}
	return nil
}
