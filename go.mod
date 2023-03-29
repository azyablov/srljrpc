module github.com/azyablov/srl-json-rpc

go 1.18

replace github.com/azyablov/srl-json-rpc/actions => ./actions

replace github.com/azyablov/srl-json-rpc/datastores => ./datastores

replace github.com/azyablov/srl-json-rpc/formats => ./formats

replace github.com/azyablov/srl-json-rpc/methods => ./methods

require (
	github.com/azyablov/srl-json-rpc/actions v0.0.0-00010101000000-000000000000
	github.com/azyablov/srl-json-rpc/datastores v0.0.0-00010101000000-000000000000
	github.com/azyablov/srl-json-rpc/formats v0.0.0-00010101000000-000000000000
	github.com/azyablov/srl-json-rpc/methods v0.0.0-00010101000000-000000000000
)
