// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package goschtalt

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/schmidtw/goschtalt/pkg/decoder"
	"github.com/schmidtw/goschtalt/pkg/encoder"
	"github.com/schmidtw/goschtalt/pkg/meta"
)

// These options must always be present to prevent panics, etc.
var alwaysOptions = []Option{
	SortRecordsNaturally(),
	AlterKeyCase(nil),
	SetKeyDelimiter("."),
}

// DefaultOptions allows a simple place where decoders can automatically register
// themselves, as well as a simple way to find what is configured by default.
// Most extensions will register themselves using init().  It is safe to change
// this value at pretty much any time & compile afterwards; just know this value
// is not mutex protected so if you are changing it after init() the synchronization
// is up to the caller.
var DefaultOptions = []Option{}

// Config is a configurable, prioritized, merging configuration registry.
type Config struct {
	mutex          sync.Mutex
	files          []string
	tree           meta.Object
	compiled       bool
	explainOptions strings.Builder
	explainCompile strings.Builder

	rawOpts []Option
	opts    options
}

// New creates a new goschtalt configuration instance with any number of options.
func New(opts ...Option) (*Config, error) {
	c := Config{
		tree: meta.Object{},
		opts: options{
			decoders: newRegistry[decoder.Decoder](),
			encoders: newRegistry[encoder.Encoder](),
		},
	}

	if err := c.With(opts...); err != nil {
		return nil, err
	}

	return &c, nil
}

// With takes a list of options and applies them.  Use of With() is optional as
// New() can take all the same options as well.  If AutoCompile() is not specified
// Compile() will need to be called to see changes in the configuration based on
// the new options.
func (c *Config) With(opts ...Option) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cfg := options{
		decoders: newRegistry[decoder.Decoder](),
		encoders: newRegistry[encoder.Encoder](),
	}

	c.explainOptions.Reset()
	c.explainCompile.Reset()

	fmt.Fprintf(&c.explainOptions, "Start of options processing.\n\n")

	raw := append(c.rawOpts, opts...)

	full := alwaysOptions

	if !ignoreDefaultOpts(raw) {
		full = append(full, DefaultOptions...)
	}

	full = append(full, c.rawOpts...)

	full = append(full, opts...)

	fmt.Fprintln(&c.explainOptions, "Options in effect:")
	i := 1
	for _, opt := range full {
		if opt != nil {
			fmt.Fprintf(&c.explainOptions, "  %d. %s\n", i, opt.String())
			i++
			if err := opt.apply(&cfg); err != nil {
				return err
			}
		}
	}

	// The options are valid, record them.
	c.opts = cfg
	c.rawOpts = raw

	fmt.Fprintf(&c.explainOptions, "\nFile extensions supported:\n")
	exts := c.opts.decoders.extensions()
	if len(exts) == 0 {
		fmt.Fprintln(&c.explainOptions, "  none")
	} else {
		for _, ext := range exts {
			fmt.Fprintf(&c.explainOptions, "  - %s\n", ext)
		}
	}

	if c.opts.autoCompile {
		if err := c.compile(); err != nil {
			return err
		}
	}

	return nil
}

// Compile reads in all the files configured using the options provided,
// and merges the configuration trees into a single map for later use.
func (c *Config) Compile() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.compile()
}

// compile is the internal compile function.
func (c *Config) compile() error {
	c.explainCompile.Reset()

	fmt.Fprintf(&c.explainCompile, "Start of compilation.\n\n")

	cfgs, err := filegroupsToRecords(c.opts.keyDelimiter, c.opts.filegroups, c.opts.decoders)
	if err != nil {
		return err
	}

	cfgs = append(cfgs, c.opts.values...)

	sorter := c.getSorter()
	sorter(cfgs)

	full := append(c.opts.defaults, cfgs...)

	merged := meta.Object{
		Map: make(map[string]meta.Object),
	}

	fmt.Fprintln(&c.explainCompile, "Records processed in order.")
	if len(full) == 0 {
		fmt.Fprintln(&c.explainCompile, "  none")
		c.tree = merged
		c.compiled = true
		return nil
	}

	files := make([]string, 0, len(full))
	for i, cfg := range full {
		fmt.Fprintf(&c.explainCompile, "  %d. %s\n", i+1, cfg.name)

		incremental := merged
		for _, exp := range c.opts.expansions {
			var err error
			incremental, err = incremental.ToExpanded(exp.maximum, exp.origin, exp.start, exp.end, exp.mapper)
			if err != nil {
				return err
			}
		}
		unmarshalFn := func(key string, result any, opts ...UnmarshalOption) error {
			// Pass in the merged value from this context and stage of processing.
			return c.unmarshal(key, result, incremental, opts...)
		}

		if err = cfg.fetch(c.opts.keyDelimiter, unmarshalFn, c.opts.decoders, c.opts.valueOptions); err != nil {
			return err
		}
		var err error
		subtree := cfg.tree.AlterKeyCase(c.opts.keySwizzler)
		merged, err = merged.Merge(subtree)
		if err != nil {
			return err
		}
		files = append(files, cfg.name)
	}

	fmt.Fprintf(&c.explainCompile, "\nVariable expansions processed in order.\n")
	if len(c.opts.expansions) == 0 {
		fmt.Fprintln(&c.explainCompile, "  none")
	}
	for i, exp := range c.opts.expansions {
		fmt.Fprintf(&c.explainCompile, "  %d. %s\n", i+1, exp.String())

		var err error
		merged, err = merged.ToExpanded(exp.maximum, exp.origin, exp.start, exp.end, exp.mapper)
		if err != nil {
			return err
		}
	}

	c.files = files
	c.tree = merged
	c.compiled = true
	return nil
}

// getSorter does the work of making a sorter for the objects we need to sort.
func (c *Config) getSorter() func([]record) {
	return func(a []record) {
		sort.SliceStable(a, func(i, j int) bool {
			return c.opts.sorter(a[i].name, a[j].name)
		})
	}
}

// ShowOrder is a helper function that provides the order the configuration
// files were combined based on the present configuration.  This can only
// be called after the Compile() has been called.
func (c *Config) ShowOrder() ([]string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.compiled {
		return []string{}, ErrNotCompiled
	}

	return c.files, nil
}

// OrderList is a helper function that sorts a caller provided list of filenames
// exectly the same way the Config object would sort them when reading and
// merging the files when the configuration is being compiled.  It also filters
// the list based on the decoders present.
func (c *Config) OrderList(list []string) (files []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cfgs := make([]record, len(list))
	for i, item := range list {
		cfgs[i] = record{name: item}
	}

	sorter := c.getSorter()
	sorter(cfgs)

	for _, cfg := range cfgs {
		file := cfg.name

		// Only include the file if there is a decoder for it.
		ext := strings.TrimPrefix(filepath.Ext(file), ".")
		_, err := c.opts.decoders.find(ext)
		if err == nil {
			files = append(files, file)
		}
	}

	return files
}

// Extensions returns the extensions this config object supports.
func (c *Config) Extensions() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.opts.decoders.extensions()
}

// Explain returns a human focused explanation of how the configuration was
// arrived at.  Each time the options change or the configuration is compiled
// the explanation will be updated.
func (c *Config) Explain() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.explainOptions.String() + "\n" + c.explainCompile.String()
}
