package bcrypt

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Hasher struct {
	options Options
}

type Options struct {
	cost int
}

func NewHasher() *Hasher {
	return NewHasherWith(Options{
		cost: 10,
	})
}

func NewHasherWith(options Options) *Hasher {
	return &Hasher{
		options: options,
	}
}

func (h *Hasher) Hash(value []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(value, h.options.cost)

	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (h *Hasher) Check(value []byte, hash []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, value)

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
