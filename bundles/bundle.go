// Package starwars provides a example schema and resolver based on Star Wars characters.
//
// Source: https://github.com/graphql/graphql.github.io/blob/source/site/_core/swapiSchema.js
package bundles

import (
	graphql "github.com/neelance/graphql-go"
)

var Schema = `
	schema {
		query: Query
		mutation: Mutation
	}
	# The query type, represents all of the entry points into our object graph
	type Query {
                all_bundles(): [Bundle]
                bundle(id: ID!): Bundle
	}
	# The mutation type, represents all updates we can make to our data
	type Mutation {
		createBundle(name: String!, path: String!): Bundle
	}
	# Represents a review for a movie
	type Bundle {
                id: ID!
		# The absolute file path where this bundles data resides
                path: String!
		# A name given to this bundle
                name: String!
	}
`

type bundle struct {
	ID        graphql.ID
	Name      string
        Path      string
}

var bundles = []*bundle{
	{
		ID: "1",
		Name: "First",
		Path: "/home/michael/bundles/First",
	},
	{
		ID: "2",
		Name: "Second",
		Path: "/home/michael/bundles/Second",
	},
}

var bundleData = make(map[graphql.ID]*bundle)

func init() {
	for _, b := range bundles {
		bundleData[b.ID] = b
	}
}

type Resolver struct{}

func (r *Resolver) All_bundles() []*bundleResolver {
	var l []*bundleResolver
	for _, bundle := range bundles {
		l = append(l, &bundleResolver{bundle})
	}
	return l
}

func (r *Resolver) Bundle(args struct{ ID graphql.ID }) *bundleResolver {
	if b := bundleData[args.ID]; b != nil {
		return &bundleResolver{b}
	}
	return nil
}


type bundleResolver struct {
	b *bundle
}

func (r *bundleResolver) ID() graphql.ID {
	return r.b.ID
}

func (r *bundleResolver) Name() string {
	return r.b.Name
}

func (r *bundleResolver) Path() string {
	return r.b.Path
}


