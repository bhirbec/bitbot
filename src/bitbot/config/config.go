/*
Package config implements loading of INI configuration file.

Example:

Assuming file "configuration.ini" contains:

	[database]
	user = emmanuel
	password = secret

You can declare configuration variables as follow:

	var (
		user     = config.String("user")
		password = config.String("password")
	)

To parse the configuration file and set a value for each variable you declared, call:

	config.Load("configuration.ini", "database")

All configuration variables are pointers:

	fmt.Printf("User is: %s", *user)
*/
package config

import (
	"fmt"

	"github.com/alyu/configparser"
)

// A Config represents a set of configuration variable.
type Config struct {
	options map[string]map[string]*string
}

// NewConfig returns a new and empty configuration.
func NewConfig() *Config {
	options := make(map[string]map[string]*string)
	return &Config{options}
}

// String declares a string variable with specified name. The returned value is
// the address of a string variable that stores the value read from the configuration file.
func (c *Config) String(section, name string) *string {
	_, ok := c.options[section]
	if !ok {
		c.options[section] = make(map[string]*string)
	}

	p := new(string)
	c.options[section][name] = p
	return p
}

// Load loads an INI file and gives a value to each variable previously declared (using
// String for instance). Must be called before variables are accessed by the program.
func (c *Config) Load(path string) error {
	rawConfig, err := configparser.Read(path)
	if err != nil {
		return fmt.Errorf("Error reading configuration file: %s", err)
	}

	for section, _ := range c.options {
		rawSection, err := rawConfig.Section(section)
		if err != nil {
			return fmt.Errorf("Error reading configuration file: %s", err)
		}

		data := rawSection.Options()

		for name, p := range c.options[section] {
			value, ok := data[name]
			if !ok {
				return fmt.Errorf("config: option %s is missing", name)
			}
			*p = value
		}
	}

	return nil
}

var (
	defaultConfig = NewConfig()

	// String wraps the String method of the default Config.
	String = defaultConfig.String

	// Load wraps the Load method of the default Config.
	Load = defaultConfig.Load
)
