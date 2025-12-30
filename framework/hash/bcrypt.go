package hash

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Bcrypt struct {
	options BcryptOptions
}

type BcryptOptions struct {
	cost int
}

const DefaultBcryptCost = 10

func NewBcrypt() *Bcrypt {
	return NewBcryptWith(BcryptOptions{
		cost: DefaultBcryptCost,
	})
}

func NewBcryptWith(options BcryptOptions) *Bcrypt {
	return &Bcrypt{
		options: options,
	}
}

func (h *Bcrypt) Hash(value []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(value, h.options.cost)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (h *Bcrypt) Check(value []byte, hash []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, value)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
