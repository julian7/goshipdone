package release_note

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"

	"gopkg.in/yaml.v3"
)

type (
	HashAlgorithm struct {
		Algo    string
		Factory func() hash.Hash
	}
)

func (h *HashAlgorithm) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("algorithm is `%v`, not a scalar!", node.Kind)
	}

	var hasherString string
	if err := node.Decode(&hasherString); err != nil {
		return fmt.Errorf("algorithm cannot be decoded: %w", err)
	}

	algo, err := NewHashAlgorithm(hasherString)
	if err != nil {
		return err
	}

	*h = *algo

	return nil
}

func NewHashAlgorithm(hasher string) (*HashAlgorithm, error) {
	factoryMap := map[string]func() hash.Hash{
		"md5":    md5.New,
		"sha1":   sha1.New,
		"sha256": sha256.New,
		"sha512": sha512.New,
	}

	factory, ok := factoryMap[hasher]
	if !ok {
		return nil, fmt.Errorf("algorithm `%s` not registered", hasher)
	}

	return &HashAlgorithm{Algo: hasher, Factory: factory}, nil
}

func (algo *HashAlgorithm) String() string {
	return algo.Algo
}
