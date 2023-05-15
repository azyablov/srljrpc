package srljrpc

import (
	"encoding/json"
	"fmt"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
)

// CommandOption type to represent a function that configures a Command.
type CommandOptions func(*Command) error

// Constructor for a new Command object with mandatory action, path and value fields, and optional command options to influence command behavior.
func NewCommand(action actions.EnumActions, path string, value CommandValue, opts ...CommandOptions) (*Command, error) {
	c := &Command{
		Path:                 path,
		Value:                string(value),
		Recursive:            nil,
		IncludeFieldDefaults: nil,
		Datastore:            &datastores.Datastore{},
	}

	if action != actions.NONE {
		c.Action = &actions.Action{}
		err := c.Action.SetAction(action)
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if opt != nil { // check that's not nil
			if err := opt(c); err != nil {
				return nil, err
			}
		}
	}
	return c, nil
}

// Provides CommandOptions to disable recursion for the command.
func WithoutRecursion() CommandOptions {
	return func(c *Command) error {
		c.withoutRecursion()
		return nil
	}
}

// CommandOptions to enable inclusion of default values in returned JSON RPC response for the command.
func WithDefaults() CommandOptions {
	return func(c *Command) error {
		c.withDefaults()
		return nil
	}
}

// CommandOptions to add path keywords to the command to substitute named parameters with the path field.
func WithAddPathKeywords(kw json.RawMessage) CommandOptions {
	return func(c *Command) error {
		return c.withPathKeywords(kw)
	}
}

// CommandOptions to set datastore for the command.
func WithDatastore(d datastores.EnumDatastores) CommandOptions {
	return func(c *Command) error {
		return c.withDatastore(d)
	}
}

// Command is mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request.
// Number of CommandOptions could be used to influence command behavior.
// Embeds Action and Datastore objects.
type Command struct {
	Path                 string          `json:"path"`
	Value                string          `json:"value,omitempty"`
	PathKeywords         json.RawMessage `json:"path-keywords,omitempty"`
	Recursive            *bool           `json:"recursive,omitempty"`
	IncludeFieldDefaults *bool           `json:"include-field-defaults,omitempty"`
	*actions.Action
	*datastores.Datastore
}

// Disable recursion for the command. Internal method.
func (c *Command) withoutRecursion() {
	v := false
	c.Recursive = &v
}

// Enable inclusion of default values in returned JSON RPC response for the command. Internal method.
func (c *Command) withDefaults() {
	v := true
	c.IncludeFieldDefaults = &v
}

// Add path keywords to the command to substitute named parameters with the path field. Internal method.
func (c *Command) withPathKeywords(jrm json.RawMessage) error {
	var data interface{}
	err := json.Unmarshal(jrm, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal path-keywords: %v", err)
	}
	c.PathKeywords = jrm
	return nil
}

// Set datastore for the command. Internal method.
func (c *Command) withDatastore(ds datastores.EnumDatastores) error {
	c.Datastore = &datastores.Datastore{}
	return c.Datastore.SetDatastore(ds)
}

// CommandValue type to represent a value of a command.
type CommandValue string

// Params defines a container for Commands and optional OutputFormat and Datastore objects.
// Params embeds OutputFormat and Datastore objects.
type Params struct {
	Commands []Command `json:"commands"`
	*formats.OutputFormat
	*datastores.Datastore
}

// Append commands to the params. Internal method.
func (p *Params) appendCommands(commands []*Command) error {
	for _, c := range commands {
		if c == nil {
			return fmt.Errorf("nil commands are not allowed")
		}
		p.Commands = append(p.Commands, *c)
	}
	return nil
}

// Set datastore for the params. Internal method.
func (p *Params) withDatastore(ds datastores.EnumDatastores) error {
	p.Datastore = &datastores.Datastore{}
	return p.Datastore.SetDatastore(ds)
}

// CLIParams defines a container for CLI commands and optional OutputFormat object.
type CLIParams struct {
	Commands []string `json:"commands"`
	*formats.OutputFormat
}

// Append commands to the params.Commands. Internal method.
func (p *CLIParams) appendCommands(commands []string) error {
	for _, c := range commands {
		if c == "" {
			return fmt.Errorf("empty commands are not allowed")
		}
	}
	p.Commands = append(p.Commands, commands...)
	return nil
}
