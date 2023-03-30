module github.com/azyablov/srljrpc

go 1.18

replace github.com/azyablov/srljrpc/actions => ./actions

replace github.com/azyablov/srljrpc/datastores => ./datastores

replace github.com/azyablov/srljrpc/formats => ./formats

replace github.com/azyablov/srljrpc/methods => ./methods

require (
	github.com/azyablov/srljrpc/actions v0.0.0-00010101000000-000000000000
	github.com/azyablov/srljrpc/datastores v0.0.0-00010101000000-000000000000
	github.com/azyablov/srljrpc/formats v0.0.0-00010101000000-000000000000
	github.com/azyablov/srljrpc/methods v0.0.0-00010101000000-000000000000
)
