package srljrpc

import (
	"encoding/json"
	"fmt"

	"github.com/azyablov/srljrpc/actions"
	"github.com/azyablov/srljrpc/datastores"
	"github.com/azyablov/srljrpc/formats"
)

// +NewCommand(EnumActions action, string path, string value, List~CommandOptions~ opts) sCommand
func NewCommand(action actions.EnumActions, path string, value CommandValue, opts ...CommandOptions) (*Command, error) {
	c := &Command{
		Path:      path,
		Recursive: true,
		Value:     string(value),
	}

	if action != actions.NONE {
		c.Action = &actions.Action{}
		err := c.Action.SetAction(action)
		if err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(c); err != nil {
				return nil, err
			}
		}
	}
	return c, nil
}

// +WithoutRecursion() CommandOption
func WithoutRecursion() CommandOptions {
	return func(c *Command) error {
		c.withoutRecursion()
		return nil
	}
}

// +WithDefaults() CommandOption
func WithDefaults() CommandOptions {
	return func(c *Command) error {
		c.withDefaults()
		return nil
	}
}

// +WithAddPathKeywords(jsonRawMessage kw) CommandOption
func WithAddPathKeywords(kw json.RawMessage) CommandOptions {
	return func(c *Command) error {
		return c.withPathKeywords(kw)
	}
}

// +WithDatastore(EnumDatastores d) CommandOption
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
	Recursive            bool            `json:"recursive,omitempty"`
	IncludeFieldDefaults bool            `json:"include-field-defaults,omitempty"`
	*actions.Action
	*datastores.Datastore
}

func (c *Command) withoutRecursion() {
	c.Recursive = false
}

func (c *Command) withDefaults() {
	c.IncludeFieldDefaults = true
}

func (c *Command) withPathKeywords(jrm json.RawMessage) error {
	var data interface{}
	err := json.Unmarshal(jrm, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal path-keywords: %v", err)
	}
	c.PathKeywords = jrm
	return nil
}

func (c *Command) withDatastore(ds datastores.EnumDatastores) error {
	c.Datastore = &datastores.Datastore{}
	return c.Datastore.SetDatastore(ds)
}

//	class CommandOption {
//		<<function>>
//		(Command c) error
//	}
type CommandOptions func(*Command) error

//	class CommandValue {
//		<<element>>
//		string
//	}
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
type Params struct {
	Commands []Command `json:"commands"`
	*formats.OutputFormat
}

func (p *Params) appendCommands(commands []*Command) {
	for _, c := range commands {
		p.Commands = append(p.Commands, *c)
	}
}

func (p *Params) getCmds() *[]Command {
	return &p.Commands
}

//	class CLIParams {
//		<<element>>
//		~List~string~ commands
//		+appendCommands(List~string~)
//	}
//
// CLIParams *-- OutputFormat
type CLIParams struct {
	Commands []string `json:"commands"`
	*formats.OutputFormat
}

func (p *CLIParams) appendCommands(commands []string) {
	p.Commands = append(p.Commands, commands...)
}

func (p *CLIParams) getCmds() *[]string {
	return &p.Commands
}
