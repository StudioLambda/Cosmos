package database

import (
	"context"

	"github.com/studiolambda/cosmos/contract"
)

type Options struct {
	table             string
	IDColumn          string
	CredentialsColumn string
	PasswordColumn    string
}

type Provider[User contract.Authenticatable] struct {
	db      contract.Database
	options Options
}

type findByIdQueryParams struct {
	table  string
	column string
	value  string
}

func NewProvider[User contract.Authenticatable](db contract.Database) *Provider[User] {
	return NewProviderWith[User](db, Options{
		IDColumn:          "id",
		CredentialsColumn: "email",
		PasswordColumn:    "password",
	})
}

func NewProviderWith[User contract.Authenticatable](db contract.Database, options Options) *Provider[User] {
	return &Provider[User]{
		db:      db,
		options: options,
	}
}

func (p *Provider[User]) UserByID(ctx context.Context, id []byte) (u User, e error) {
	const q = `select * from $1 where $2=$3 limit 1`

	if err := p.db.Find(ctx, q, &u, p.options.table, p.options.IDColumn, id); err != nil {
		return u, err
	}

	return u, e
}

func (p *Provider[User]) RetrieveByCredentials(ctx context.Context, identifier []byte) (u User, e error) {
	const q = `select * from $1 where $2=$3 limit 1`

	if err := p.db.Find(ctx, q, &u, p.options.table, p.options.CredentialsColumn, identifier); err != nil {
		return u, err
	}

	return u, e
}

func (p *Provider[User]) Validate(ctx context.Context, user User, password []byte) bool {
	//

	return false
}
