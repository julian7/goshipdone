package modules

import (
	"compress/gzip"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type (
	// Compressor defines compression interface
	Compressor interface {
		fmt.Stringer
		Extension() string
		Writer(io.Writer) io.WriteCloser
	}

	// Compression is a YAML representation of a compression format
	Compression struct {
		Compressor
	}

	// CompressNONE defines a flowthrough compression
	CompressNONE struct{}

	// CompressGz defines a gzip compression
	CompressGz struct{}

	nopWriteCloser struct {
		io.Writer
	}
)

// UnmarshalYAML detects compression format
func (c *Compression) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("compression is `%v`, not scalar", node.Kind)
	}

	var compressString string
	if err := node.Decode(&compressString); err != nil {
		return fmt.Errorf("compression cannot be decoded: %w", err)
	}

	switch compressString {
	case "", "none", "NONE":
		(*c) = Compression{&CompressNONE{}}
	case "gz", "gzip", "GZip":
		(*c) = Compression{&CompressGz{}}
	default:
		return fmt.Errorf("invalid compression format: `%s`", compressString)
	}

	return nil
}

func (c *CompressNONE) String() string {
	return "NONE"
}

func (c *CompressNONE) Extension() string {
	return ""
}

func (c *CompressNONE) Writer(writer io.Writer) io.WriteCloser {
	return &nopWriteCloser{Writer: writer}
}

func (c *CompressGz) String() string {
	return "gzip"
}

func (c *CompressGz) Extension() string {
	return ".gz"
}

func (c *CompressGz) Writer(writer io.Writer) io.WriteCloser {
	return gzip.NewWriter(writer)
}

func (nopWriteCloser) Close() error {
	return nil
}
