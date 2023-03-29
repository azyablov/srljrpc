```mermaid
---
title: SR Linux JSON RPC
---
classDiagram
    class EnumMethods {
        <<enumeration>>
        note "Used to retrieve configuration and state details from the system. The get method can be used with candidate, running, and state datastores, but cannot be used with the tools datastore." 
        GET
        note "Used to set a configuration or run operational transaction. The set method can be used with the candidate and tools datastores."
        SET
        note "Used to run CLI commands. The get and set methods are restricted to accessing data structures via the YANG models, but the cli method can access any commands added to the system via python plug-ins or aliases."
        CLI
        note "Used to verify that the system accepts a configuration transaction before applying it to the system."
        VALIDATE
    }
    EnumMethods "1" --o Method: is one of

    class EnumOutputFormats {
        <<enumeration>>
        JSON
        TEXT
        TABLE
    }
    EnumOutputFormats "1" --o OutputFormat: OneOf

    class EnumActions {
        <<enumeration>>
        note "Replaces the entire configuration within a specific context with the supplied configuration; equivalent to a delete/update. When the action command is used with the tools datastore, update is the only supported option."
        REPLACE
        note "Updates a leaf or container with the specified value." 
        UPDATE
        note "Deletes a leaf or container. All children beneath the parent are removed from the system."
        DELETE
    }
    EnumActions "1" --o Action: OneOf

    class EnumDatastores {
        <<enumeration>>
        note "Used to change the configuration of the system with the get, set, and validate methods; default datastore is used if the datastore parameter is not provided."
        CANDIDATE
        note "Used to retrieve the active configuration with the get method."
        RUNNING
        note "Used to retrieve the running (active) configuration along with the operational state."
        STATE
        note "Used to perform operational tasks on the system; only supported with the update action command and the set method."
        TOOLS
    }
    EnumDatastores "1" --o Datastore: OneOf

    note for action "Conditional mandatory; used with the set and validate methods."
    class Action {
        <<element>>
        ~GetAction(): EnumActions
        ~SetAction(a: EnumActions): error
        +string Action
    }

    note for datastore "Optional; selects the datastore to perform the method against. CANDIDATE datastore is used if the datastore parameter is not provided."
    class Datastore {
        <<element>>
        +GetDatastore(): EnumDatastores
        +SetDatastore(d: EnumDatastores): error
        +string Datastore
    }

    note for Command "Mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request."
    class Command {
        <<element>>
        note "Mandatory with the get, set and validate methods. This value is a string that follows the gNMI path specification1 in human-readable format."
        ~string Path
        note "Optional, since can be embedded into path, for such kind of cases value should not be specified, so path assumed to follow <path>:<value> schema, which will be checked for set and validate"
        ~string Value
        note "Optional; used to substitute named parameters with the path field. More than one keyword can be used with each path."
        ~string PathKeywords
        note "Optional; a Boolean used to retrieve children underneath the specific path. The default = true."
        ~bool Recursive
        note "Optional; a Boolean used to show all fields, regardless if they have a directory configured or are operating at their default setting. The default = false."
        ~bool Include-field-defaults
        ~withoutRecursion()
        ~withDefaults()
        ~withPathKeywords(jsonRawMessage) error
        ~withDatastore(EnumDatastores)
    }
    Command *-- "1" Action
    Command *-- "1" Datastore
    
    note for outputFormat "Optional. Defines the output format. Output defaults to JSON if not specified."
    class OutputFormat {
        <<element>>
        +GetFormat() EnumOutputFormats
        +SetFormat(EnumOutputFormats of) error
        #string OutputFormat
    }

    note for params "MAY be omitted. Defines a container for any parameters related to the request. The type of parameter is dependent on the method used."
    class Params {
        <<element>>
        ~List~Command~ Commands
        +appendCommands(List~Command~)
        ~getCmds() List~Command~
    }
    Params *-- OutputFormat

    class CLIParams {
        <<element>>
        ~List~string~ commands
        +appendCommands(List~string~)
    }
    CLIParams *-- OutputFormat
    
    
    note for method "Mandatory. Supported options are get, set, and validate. "
    class Method {
        <<element>>
        ~GetMethod() EnumMethods
        ~SetMethod(EnumMethods) bool
        #string Method
    }

    note for Request "JSON RPC Request: get / set / validate"
    class Request {
        <<message>>
        note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
        ~string JSONRpcVersion
        note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
        ~int ID
        +Marshal() List~byte~
        +GetID() int
    }
    Request *-- Method
    Request *-- Params

    class Requester {
        <<interface>>
        +GetMethod() string
        +Marshal() List~byte~
        +GetID() int
    }

    note for CLIRequest "JSON RPC Request: cli"
    class CLIRequest {
        <<message>>
        note "Method set to CLI"
        note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
        ~string JSONRpcVersion
        note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
        ~int ID
        note "Mandatory. Supported options are cli. Set statically in the RPC request"
        +Marshal() List~byte~
        +GetID() int
        ~setID(int)
    }
    CLIRequest *-- Method
    CLIRequest *-- CLIParams

    note for RpcError "When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response is expressed as a single JSON Object"
    class RpcError {
        <<element>>
        note "A Number that indicates the error type that occurred. This MUST be an integer."
        +int ID
        note "A String providing a short description of the error. The message SHOULD be limited to a concise single sentence."
        +string Message
        note "A Primitive or Structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.)."
        +string Data
    }

    class Response {
        <<message>>
        note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
        ~string JSONRpcVersion
        note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
        ~int ID
        note "This member is REQUIRED on success. This member MUST NOT exist if there was an error invoking the method. The value of this member is determined by the method invoked on the Server."
        +jsonRawMessage Result
        note "This member is REQUIRED on error. This member MUST NOT exist if there was no error triggered during invocation. The value for this member MUST be an Object as defined in section 5.1."
        +RpcError Error
    }
    Response o-- RpcError

    class JSONRPCClient {
        <<entity>>
        Call(Requester r) Response
    }

    class CommandValue {
        <<element>>
        string
    }

    class jsonrpc {
        <<module>>
        +NewJSONRPCClient(SRLTarget t) JSONRPCClient

        +NewCLIRequest(List~string~ cmds, List~RequestOption~ opts) CLIRequest
        +NewRequest(EnumMethods m, List~GetCommand~ cmds, List~RequestOption~ opts) Request


        +WithOutputFormat(EnumOutputFormats of) RequestOption

        %% Commands
        +NewCommand(EnumActions action, string path, CommandValue value, List~CommandOptions~ opts) Command
        
        %% Command options for GET, SET, VALIDATE
        +WithoutRecursion() CommandOption
        +WithDefaults() CommandOption
        +WithAddPathKeywords(jsonRawMessage kw) CommandOption
        +WithDatastore(EnumDatastores d) CommandOption
    }

    class CommandOption {
        <<function>>
        (Command c) error
    }

    class RequestOption {
        <<function>>
        (Requester c) error
    }
```