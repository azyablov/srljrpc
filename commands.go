package srljrpc

import (
	"encoding/json"
	"fmt"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
)

// +NewCommand(EnumActions action, string path, string value, List~CommandOptions~ opts) Command
// Constructor for Command object with mandatory action, path and value fields, and optional command options to influence command behaviour.
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

// +WithoutRecursion() CommandOption
// Disable recursion for the command.
func WithoutRecursion() CommandOptions {
	return func(c *Command) error {
		c.withoutRecursion()
		return nil
	}
}

// +WithDefaults() CommandOption
// Enable inclusion of default values in returned JSON RPC response for the command.
func WithDefaults() CommandOptions {
	return func(c *Command) error {
		c.withDefaults()
		return nil
	}
}

// +WithAddPathKeywords(jsonRawMessage kw) CommandOption
// Add path keywords to the command to substitute named parameters with the path field.
func WithAddPathKeywords(kw json.RawMessage) CommandOptions {
	return func(c *Command) error {
		return c.withPathKeywords(kw)
	}
}

// +WithDatastore(EnumDatastores d) CommandOption
// Set datastore for the command.
func WithDatastore(d datastores.EnumDatastores) CommandOptions {
	return func(c *Command) error {
		return c.withDatastore(d)
	}
}

// note for Command "Mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request."
//
//	class Command {
//		<<element>>
//		note "Mandatory with the get, set and validate methods. This value is a string that follows the gNMI path specification1 in human-readable format."
//		~string Path
//		note "Optional, since can be embedded into path, for such kind of cases value should not be specified, so path assumed to follow <path>:<value> schema, which will be checked for set and validate"
//		~string Value
//		note "Optional; used to substitute named parameters with the path field. More than one keyword can be used with each path."
//		~string PathKeywords
//		note "Optional; a Boolean used to retrieve children underneath the specific path. The default = true."
//		~bool Recursive
//		note "Optional; a Boolean used to show all fields, regardless if they have a directory configured or are operating at their default setting. The default = false."
//		~bool Include-field-defaults
//		+withoutRecursion()
//		+withDefaults()
//		+withPathKeywords(jsonRawMessage) error
//		+withDatastore(EnumDatastores)
//		+GetDatastore() string
//	}
//
// Command *-- "1" Action
// Command *-- "1" Datastore
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

//	class CommandOption {
//		<<function>>
//		(Command c) error
//	}
//
// CommandOption type to represent a function that configures a Command.
type CommandOptions func(*Command) error

//	class CommandValue {
//		<<element>>
//		string
//	}
//
// CommandValue type to represent a value of a command.
type CommandValue string

// note for params "MAY be omitted. Defines a container for any parameters related to the request. The type of parameter is dependent on the method used."
//
//	class Params {
//		<<element>>
//		~List~Command~ commands
//		+appendCommands(List~Command~)
//	}
//
// Params *-- OutputFormat
// Params class implementation.
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

// To be reconsidered but looks as redundant for now.
// func (p *Params) getCmds() *[]Command {
// 	return &p.Commands
// }

// Set datastore for the params. Internal method.
func (p *Params) withDatastore(ds datastores.EnumDatastores) error {
	p.Datastore = &datastores.Datastore{}
	return p.Datastore.SetDatastore(ds)
}

//	class CLIParams {
//		<<element>>
//		~List~string~ commands
//		+appendCommands(List~string~)
//	}
//
// CLIParams *-- OutputFormat
// CLIParams class implementation.
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

// To be reconsidered but looks as redundant for now.
// func (p *CLIParams) getCmds() *[]string {
// 	return &p.Commands
// }
