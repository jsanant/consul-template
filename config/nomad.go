package config

import "net/http"

type NomadConfig struct {
	Address      *string      `mapstructure:"address"`
	Namespace    *string      `mapstructure:"namespace"`
	Token        *string      `mapstructure:"token"`
	AuthUsername *string      `mapstructure:"auth_username"`
	AuthPassword *string      `mapstructure:"auth_password"`
	HttpClient   *http.Client `mapstructure:"http_client"`
}

func DefaultNomadConfig() *NomadConfig {
	return &NomadConfig{}
}

// Copy returns a deep copy of this configuration.
func (n *NomadConfig) Copy() *NomadConfig {
	if n == nil {
		return nil
	}

	var o NomadConfig

	o.Address = n.Address
	o.Namespace = n.Namespace
	o.Token = n.Token
	o.AuthUsername = n.AuthUsername
	o.AuthPassword = n.AuthPassword
	o.HttpClient = n.HttpClient

	return &o
}

// Merge combines all values in this configuration with the values in the other
// configuration, with values in the other configuration taking precedence.
// Maps and slices are merged, most other values are overwritten. Complex
// structs define their own merge functionality.
func (n *NomadConfig) Merge(o *NomadConfig) *NomadConfig {
	if n == nil {
		if o == nil {
			return nil
		}
		return o.Copy()
	}

	if o == nil {
		return n.Copy()
	}

	r := n.Copy()

	if o.Address != nil {
		r.Address = o.Address
	}

	if o.Namespace != nil {
		r.Namespace = o.Namespace
	}

	if o.Token != nil {
		r.Token = o.Token
	}

	if o.AuthUsername != nil {
		r.AuthUsername = o.AuthUsername
	}

	if o.AuthPassword != nil {
		r.AuthPassword = o.AuthPassword
	}

	if o.HttpClient != nil {
		r.HttpClient = o.HttpClient
	}

	return r
}

// Finalize ensures there no nil pointers.
func (n *NomadConfig) Finalize() {

	if n.Address == nil {
		n.Address = stringFromEnv([]string{"NOMAD_ADDR"}, "")
	}

	if n.Namespace == nil {
		n.Namespace = stringFromEnv([]string{"NOMAD_NAMESPACE"}, "")
	}

	if n.Token == nil {
		n.Token = stringFromEnv([]string{"NOMAD_TOKEN"}, "")
	}

	if n.AuthUsername == nil {
		n.AuthUsername = String("")
	}

	if n.AuthPassword == nil {
		n.AuthPassword = String("")
	}

	if n.HttpClient == nil {
		n.HttpClient = http.DefaultClient
	}
}
