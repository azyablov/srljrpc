# SR Linux JSON RPC client
JSON RPC client library implementation for SR Linux aimed to simplify the ways talking to SR Linux devices via JSON RPC.

# How To

## Introduction

Besides just self explanatory naming it's good to create a sort of the workflow to describe semantics of provided API. 
Giving some outlook on available methods and options is quite important as well, 
while working on the first implementation in GO is was not so obvious how provide simple interface at the same time 
hiding overall complexity via the exposure of well understandable interface elements.
This document try to explore available without sophisticated scenarios like service activation, but rather focus on elementary steps and actions to build them.

## API workflow

As usual for the sake of demo, we will use [Containerlab][clab] as brilliant network simulation tool and one of the standard labs published by SRL Labs [EVPN interoperability between SR Linux and SROS][evpn].
How to bring this is up and used very well described while you following provided links.
Our simple program will grow up each and every steps allowing to to have necessary grip with the subject API.

### Client creation

As you can see we have very simple code to start with JSON RPC client. While client is created necessary pre-flight checks are done in background to elicit system hostname and software version.
That serving two objectives: verifying immediately and get HTTP client ready, giving minimum information about system in order to take decision about necessary YANG modules to use further if you would like to support ENV with mix of SR Linux versions.

```golang
package main

import (
	"encoding/json"
	"fmt"

	"github.com/azyablov/srljrpc"
	"github.com/azyablov/srljrpc/formats"
)


var (
	host = "clab-evpn-leaf1"
	user = "admin"
	pass = "admin"
	port = 443
)

func main() {
	// Create a new JSON RPC client with credentials and port (used 443 as default for the sake of demo).
	c, err := srljrpc.NewJSONRPCClient(&host, srljrpc.WithOptCredentials(&user, &pass), srljrpc.WithOptPort(&port))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Target hostname: %s\nTarget system version: %s\n", c.GetHostname(), c.GetSysVer())
}
```

Finally, our simple program is printing hostname and system version.

```sh
[azyablov@ecartman srljrpc_client_example]$ go run client.go 
Target hostname: leaf1
Target system version: v22.6.2-24-g5e9fff1e5b
[azyablov@ecartman srljrpc_client_example]$ 
```

#### Options available

Worth to mention that JSON RPC client has a few options available:
- ```WithOptPort(port *int)```
- ```WithOptTimeout(t time.Duration)```
- ```WithOptCredentials(u, p *string)```
- ```WithOptTLS(t *TLSAttr)```

All of them are quite self-descriptive, but ```WithOptTLS``` should be a bit more explained to give 100% confidence.
First of all, JSON file to TLSAttr object looks like the following (taken from real lab):
```json
    {
        "tls_attr": {
            "skip_verify": false,
            "cert_file": "/home/azyablov/clab/nokia-evpn-lab/clab-evpn/ca/spine1/spine1.pem",
            "key_file": "/home/azyablov/clab/nokia-evpn-lab/clab-evpn/ca/spine1/spine1-key.pem",
            "ca_file": "/home/azyablov/clab/nokia-evpn-lab/clab-evpn/ca/root/root-ca.pem"
        }
    }
```

The last could be read from file / string / everything implements Read interface, so basically nothing new:

```golang
    var ta TLSAttr
	err := json.Unmarshal([]byte(jsonStr), &ta)
	if err != nil {
		panic(err)
	}
```

### Sending requests
#### Getting config 

Now, let's read something from running configuration by using the next two xpaths: 
- ```/network-instance[name="MAC-VRF 1"]```
- ```/system/lldp```

As you can see code is quite trivial:

```golang
	// GET method example.
	getResp, err := c.Get(`/network-instance[name="MAC-VRF 1"]`, `/system/lldp`)
	if err != nil {
		panic(err)
	}
	rStr, err := json.MarshalIndent(getResp.Result, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response from GET: %s\n", string(rStr))
```

As soon as we submitted two xpath, we are getting two elements in the list of ```getResp.Result```:

```json
Target hostname: leaf1
Target system version: v22.6.2-24-g5e9fff1e5b
c.Get() example:
Response: [
  {
    "type": "srl_nokia-network-instance:mac-vrf",
    "interface": [
      {
        "name": "ethernet-1/1.1"
      }
    ],
    "vxlan-interface": [
      {
        "name": "vxlan0.1"
      }
    ],
    "protocols": {
      "bgp-evpn": {
        "srl_nokia-bgp-evpn:bgp-instance": [
          {
            "id": 1,
            "admin-state": "enable",
            "vxlan-interface": "vxlan0.1",
            "evi": 1,
            "ecmp": 2
          }
        ]
      },
      "srl_nokia-bgp-vpn:bgp-vpn": {
        "bgp-instance": [
          {
            "id": 1,
            "route-distinguisher": {
              "rd": "1:11"
            },
            "route-target": {
              "export-rt": "target:65011:1",
              "import-rt": "target:65011:1"
            }
          }
        ]
      }
    }
  },
  {
    "admin-state": "enable"
  }
]
```

#### Getting state 

Let's image we have to have some stats/operational state alongside with configuration info, so you need to use STATE datastore in order to get it.
Well, one shoot of Stats() methods is resolving it in quite convenient way.
In the example below we are using ```/system/json-rpc-server``` xpath.
```golang
    // Getting stats.
	fmt.Println("c.State() example:")
	stateResp, err := c.State("/system/json-rpc-server")
	if err != nil {
		panic(err)
	}
	outHelper(stateResp.Result)
```
In order to read code more readable outHelper() function was introduced, which is unmarshalling and printing results:

```golang
func outHelper(v any) {
	rStr, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(rStr))
}
```

So, you should see something more on top of already mentioned output.

```json
c.State() example:
[
  {
    "admin-state": "enable",
    "commit-confirmed-timeout": 0,
    "network-instance": [
      {
        "name": "mgmt",
        "http": {
          "admin-state": "enable",
          "oper-state": "up",
          "use-authentication": true,
          "session-limit": 10,
          "port": 80
        },
        "https": {
          "admin-state": "enable",
          "oper-state": "up",
          "use-authentication": true,
          "session-limit": 10,
          "port": 443,
          "tls-profile": "clab-profile"
        }
      }
    ],
    "unix-socket": {
      "admin-state": "disable",
      "oper-state": "down",
      "use-authentication": true,
      "socket-path": ""
    }
  }
]

```

#### Updating/Replacing/Deleting config

Example below is reading values before UPDATE/DELETE/REPLACE operations and executing them respectively.
```Validate()``` and ```Tools()``` methods are available as well, 
first is used to validate configuration updates w/o applying it, the second one used to perform /tools operations like ```/tools interface ethernet-1/1 statistics clear```, 
but they aren't giving anything special semantics in terms of API interaction and omitted for brevity.

```golang
	// Updating/Replacing/Deleting config
	fmt.Println("c.Update()/Delete()/Replace() example:")
	pvs := []srljrpc.PV{
		{Path: `/interface[name=ethernet-1/51]/subinterface[index=0]/description`, Value: "UPDATE"},
		{Path: `/system/banner`, Value: "DELETE"},
		{Path: `/interface[name=mgmt0]/description`, Value: "REPLACE"},
	}
	// Getting existing config for the sake of demo.
	for _, pv := range pvs {
		getResp, err := c.Get(pv.Path)
		if err != nil {
			panic(err)
		}
		outHelper(getResp.Result)
	}

	mdmResp, err := c.Update(pvs[0])
	if err != nil {
		panic(err)
	}
	outHelper(mdmResp.Result)
	mdmResp, err = c.Delete(pvs[1].Path)
	if err != nil {
		panic(err)
	}
	outHelper(mdmResp.Result)
	mdmResp, err = c.Replace(pvs[2])
	if err != nil {
		panic(err)
	}
	outHelper(mdmResp.Result)
```

Essentially one operation is just four lines of code:

```golang
	mdmResp, err = c.Replace(pvs[2])
	if err != nil {
		panic(err)
	}
```

Console output should be altered by the following contents:

```json
.Update()/Delete()/Replace() example:
[
  "to_spine1"
]
[
  {
    "login-banner": "................................................................\n:                  Welcome to Nokia SR Linux!                  :\n:              Open Network OS for the NetOps era.             :\n:                                                              :\n:    This is a freely distributed official container image.    :\n:                      Use it - Share it                       :\n:                                                              :\n: Get started: https://learn.srlinux.dev                       :\n: Container:   https://go.srlinux.dev/container-image          :\n: Docs:        https://doc.srlinux.dev/22-6                    :\n: Rel. notes:  https://doc.srlinux.dev/rn22-6-2                :\n: YANG:        https://yang.srlinux.dev/v22.6.2                :\n: Discord:     https://go.srlinux.dev/discord                  :\n: Contact:     https://go.srlinux.dev/contact-sales            :\n................................................................\n"
  }
]
[
  {}
]
[
  {}
]
[
  {}
]
[
  {}
]
```

If we would again query it we should get the following output, since commit was applied before (automatically by JSON RPC servers) and running configuration updated:

```json
[
  "UPDATE"
]
[
  {}
]
[
  "REPLACE"
]
```


### Sending CLI commands

Sending CLI commands is one of the main methods to interact with network devices, even industry is rapidly adopting MDM interfaces.
Lab builds, validations, troubleshooting and many other operational tasks would require interaction with CLI.
JSON RPC interface of SR Linux is providing very convenient way to automate and use CLI commands, 
especially we you are using number of CLI plug-ins to elicit essential information, but still giving you full flexibility to utilize structured data outputs.

#### Executing CLI commands


In CLI example below two commands are executed with output format JSON and one with out format TABLE.
For the output format TABLE ```[]string``` type were used to marshal it correctly into string and print in nice form in STDOUT.
As easy to see number of line of code remains +/- stable in terms of getting necessary results. Of course, assuming logic of your application is not counted here.

```golang
	// CLI
	fmt.Println("c.CLI() example:")
	cliResp, err := c.CLI([]string{"show version", "show network-instance summary"}, formats.JSON)
	if err != nil {
		panic(err)
	}
	outHelper(cliResp.Result)

	cliResp, err = c.CLI([]string{"show system lldp neighbor"}, formats.TABLE)
	if err != nil {
		panic(err)
	}
	type Table []string
	var t Table
	b, _ := cliResp.Result.MarshalJSON()
	err = json.Unmarshal(b, &t)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", t[0])
```

```json
c.CLI() example:
[
  {
    "basic system info": {
      "Hostname": "leaf1",
      "Chassis Type": "7220 IXR-D2",
      "Part Number": "Sim Part No.",
      "Serial Number": "Sim Serial No.",
      "System HW MAC Address": "1A:B8:02:FF:00:00",
      "Software Version": "v22.6.2",
      "Build Number": "24-g5e9fff1e5b",
      "Architecture": "x86_64",
      "Last Booted": "2023-04-23T18:42:30.977Z",
      "Total Memory": "23640339 kB",
      "Free Memory": "7673355 kB"
    }
  },
  {
    "Network Instance": [
      {
        "Name": "MAC-VRF 1",
        "Type": "mac-vrf",
        "Admin state": "enable",
        "Oper state": "up",
        "Router id": "N/A"
      },
      {
        "Name": "default",
        "Type": "default",
        "Admin state": "enable",
        "Oper state": "up"
      },
      {
        "Name": "mgmt",
        "Type": "ip-vrf",
        "Admin state": "enable",
        "Oper state": "up",
        "Description": "Management network instance"
      }
    ]
  }
]
```
```sh
+---------------------------+---------------------------+---------------------------+---------------------------+---------------------------+---------------------------+---------------------------+
|           Name            |         Neighbor          |   Neighbor System Name    |    Neighbor Chassis ID    |  Neighbor First Message   |   Neighbor Last Update    |       Neighbor Port       |
+===========================+===========================+===========================+===========================+===========================+===========================+===========================+
| ethernet-1/51             | 1A:35:06:FF:00:00         | spine1                    | 1A:35:06:FF:00:00         | 3 hours ago               | now                       | ethernet-1/11             |
| ethernet-1/52             | 1A:CA:07:FF:00:00         | spine2                    | 1A:CA:07:FF:00:00         | 3 hours ago               | now                       | ethernet-1/11             |
| mgmt0                     | 1A:0F:04:FF:00:00         | leaf3                     | 1A:0F:04:FF:00:00         | 3 hours ago               | now                       | mgmt0                     |
| mgmt0                     | 1A:35:06:FF:00:00         | spine1                    | 1A:35:06:FF:00:00         | 3 hours ago               | now                       | mgmt0                     |
| mgmt0                     | 1A:95:03:FF:00:00         | leaf2                     | 1A:95:03:FF:00:00         | 3 hours ago               | now                       | mgmt0                     |
| mgmt0                     | 1A:B4:05:FF:00:00         | leaf4                     | 1A:B4:05:FF:00:00         | 3 hours ago               | now                       | mgmt0                     |
| mgmt0                     | 1A:CA:07:FF:00:00         | spine2                    | 1A:CA:07:FF:00:00         | 3 hours ago               | now                       | mgmt0                     |
+---------------------------+---------------------------+---------------------------+---------------------------+---------------------------+---------------------------+---------------------------+
```

All examples provided in this document can be found in [repository][samples] with SR Linux JSON RPC library samples.


[clab]: https://containerlab.dev
[evpn]: https://github.com/srl-labs/nokia-evpn-lab
[samples]: https://github.com/azyablov/srljrpc_client_example