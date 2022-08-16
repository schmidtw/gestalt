// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const redactedText = "REDACTED"

const (
	cmdSecret  = "secret"
	cmdReplace = "replace"
	cmdKeep    = "keep"
	cmdFail    = "fail"
	cmdAppend  = "append"
	cmdPrepend = "prepend"
	cmdSplice  = "splice"
	cmdClear   = "clear"
)

var (
	ErrConflict         = errors.New("a conflict has been detected")
	ErrInvalidCommand   = errors.New("invalid command")
	ErrNotFound         = errors.New("not found")
	ErrArrayOutOfBounds = errors.New("array index is out of bounds")
)

// Origin provides details about an origin of a parameter.
type Origin struct {
	File string // Filename where the value originated.
	Line int    // Line number where the value originated.
	Col  int    // Column where the value originated.
}

// String returns a useful representation for the origin.
func (o Origin) String() string {
	file := "unknown"
	line := "???"
	col := "???"

	if len(o.File) > 0 {
		file = o.File
	}
	if o.Line > 0 {
		line = strconv.Itoa(o.Line)
	}
	if o.Col > 0 {
		col = strconv.Itoa(o.Col)
	}
	return fmt.Sprintf("%s:%s[%s]", file, line, col)
}

const (
	Array = iota + 1
	Map
	Value
)

// Object represents either a map, of objects, and array of objects or a specific
// configuration value and the origins of how this Object came into existence.
//
// To build an Object tree, either add to the Array, Map or Value fields.  Don't
// add to them all, as only one will be used.  The order is Array > Map > Value.
// So for example, if you add to the Array field and the Value field, the Value
// field will always be ignored.
type Object struct {
	Origins []Origin          // The list of origins that influenced this Object.
	Array   []Object          // The array of Objects (if a map).
	Map     map[string]Object // The map of Objects (if a map).
	Value   any               // The value of the configuration parameter (if a value).
	secret  bool              // If the value is secret.
}

// Kind provides the specific kind of Object this is.  Array, Map or Value.  If
// it unclear exactly which, Value will be returned.
func (o Object) Kind() int {
	if 0 < len(o.Array) {
		return Array
	}
	if 0 < len(o.Map) {
		return Map
	}
	return Value
}

// OriginString provides the string for all origins for this Object.
func (obj Object) OriginString() string {
	var list []string
	for _, v := range obj.Origins {
		list = append(list, v.String())
	}

	return strings.Join(list, ", ")
}

// Fetch looks up the specific asks in the tree (map keys or array indexes) and
// returns the found object or provides a contextual error.  The separater is
// used to provide error context.
func (obj Object) Fetch(asks []string, separater string) (Object, error) {
	return obj.fetch(asks, asks, separater)
}

// getPath is an internal helper that determines the path in use.  Mainly used
// for determining the portion of the original string where the error was
// encountered at.
func getPath(asks, path []string, separater string) string {
	// Trim off the same number of elements that are left in the asks from the
	// path before joining them and returning them.
	path = path[:len(path)-len(asks)]
	return strings.Join(path, separater)
}

// fetch is the internal helper function that actually finds and returns the
// Object of interest.
func (obj Object) fetch(asks, path []string, separater string) (Object, error) {
	if len(asks) == 0 {
		return obj, nil
	}

	switch obj.Kind() {
	case Map:
		key := asks[0]
		next, found := obj.Map[key]
		if found {
			return next.fetch(asks[1:], path, separater)
		}
	case Array:
		idx, err := strconv.Atoi(asks[0])
		if err != nil {
			return Object{}, err
		}
		if 0 <= idx && idx < len(obj.Array) {
			return obj.Array[idx].fetch(asks[1:], path, separater)
		} else {
			return Object{},
				fmt.Errorf("with array len of %d and '%s' %w",
					len(obj.Array),
					getPath(asks[1:], path, separater),
					ErrArrayOutOfBounds)
		}
	}

	return Object{}, fmt.Errorf("with '%s' %w",
		getPath(asks[1:], path, separater), ErrNotFound)
}

// ToRaw converts an Object tree into a native go tree (with no secret or origin history.
func (obj Object) ToRaw() any {
	switch obj.Kind() {
	case Array:
		rv := make([]any, len(obj.Array))
		for i, val := range obj.Array {
			rv[i] = val.ToRaw()
		}
		return rv
	case Map:
		rv := make(map[string]any)

		for key, val := range obj.Map {
			rv[key] = val.ToRaw()
		}
		return rv
	}
	return obj.Value
}

// ObjectFromRaw converts a native go tree into the equivalent Object tree structure.
func ObjectFromRaw(in any) (obj Object) {
	obj.Origins = []Origin{}

	switch in := in.(type) {
	case []any:
		obj.Array = make([]Object, len(in))
		for i, val := range in {
			obj.Array[i] = ObjectFromRaw(val)
		}
	case map[string]any:
		obj.Map = make(map[string]Object)
		for key, val := range in {
			obj.Map[key] = ObjectFromRaw(val)
		}
	default:
		obj.Value = in
	}

	return obj
}

// ToRedacted builds a copy of the tree where secrets are redacted.  Secret maps
// or arrays will now show up as values containing the value 'REDACTED'.
func (obj Object) ToRedacted() Object {
	if obj.secret {
		return Object{
			Origins: []Origin{},
			Value:   redactedText,
			secret:  true,
		}
	}

	switch obj.Kind() {
	case Array:
		array := make([]Object, len(obj.Array))
		for i, val := range obj.Array {
			array[i] = val.ToRedacted()
		}
		obj.Array = array
	case Map:
		m := make(map[string]Object)

		for key, val := range obj.Map {
			m[key] = val.ToRedacted()
		}
		obj.Map = m
	}

	return obj
}

// AlterKeyCase builds a copy of the tree where the keys for all Objects have
// been converted using the specified conversion function.
func (obj Object) AlterKeyCase(to func(string) string) Object {
	switch obj.Kind() {
	case Array:
		array := make([]Object, len(obj.Array))
		for i, val := range obj.Array {
			array[i] = val.AlterKeyCase(to)
		}
		obj.Array = array
	case Map:
		m := make(map[string]Object)

		for key, val := range obj.Map {
			m[to(key)] = val.AlterKeyCase(to)
		}
		obj.Map = m
	}

	return obj
}

// ResolveCommands builds a copy of the tree where the commands have been
// resolved from the keys.
func (obj Object) ResolveCommands() (Object, error) {
	return obj.resolveCommands(false)
}

// resolveCommands is the internal helper function that does the actual resolution.
func (obj Object) resolveCommands(secret bool) (Object, error) {
	if secret {
		obj.secret = true
	}

	switch obj.Kind() {
	case Array:
		array := make([]Object, len(obj.Array))
		for i, val := range obj.Array {
			v, err := val.resolveCommands(false)
			if err != nil {
				return Object{}, err
			}
			array[i] = v
		}
		obj.Array = array
	case Map:
		m := make(map[string]Object)

		for key, val := range obj.Map {
			cmd, err := getValidCmd(key, val)
			if err != nil {
				return Object{}, err
			}
			tmp, err := val.resolveCommands(cmd.secret)
			if err != nil {
				return Object{}, err
			}
			m[cmd.final] = tmp
		}
		obj.Map = m
	}

	return obj, nil
}

// Merge performs a merge of the new Object tree onto the existing Object tree
// using the default semantics and merge rules found in the key commands.
func (obj Object) Merge(next Object) (Object, error) {
	// The 'clear' command is special in that if it is found at all, it
	// overwrites everything else in the existing tree and exists the merge.
	for k := range next.Map {
		cmd, err := getCmd(k)
		if err != nil {
			return Object{}, err
		}
		if cmd.cmd == cmdClear {
			return Object{Origins: []Origin{}}, nil
		}
	}

	return obj.merge(command{}, next)
}

// merge does the actual merging of the trees.
func (obj Object) merge(cmd command, next Object) (Object, error) {
	kind := obj.Kind()
	if kind == Value {
		rv := obj
		switch cmd.cmd {
		case cmdReplace, "":
			rv = next
		case cmdKeep:
		case cmdFail:
			return Object{}, ErrConflict
		}

		rv.secret = cmd.secret
		return rv, nil
	}

	if kind == Array {
		rv := obj
		next, err := next.resolveCommands(obj.secret)
		if err != nil {
			return Object{}, err
		}
		switch cmd.cmd {
		case cmdAppend, "":
			if obj.secret || next.secret || cmd.secret {
				rv.secret = true
			}
			rv.Origins = append(obj.Origins, next.Origins...)
			rv.Array = append(obj.Array, next.Array...)
		case cmdPrepend:
			if obj.secret || next.secret || cmd.secret {
				rv.secret = true
			}
			rv.Origins = append(next.Origins, obj.Origins...)
			rv.Array = append(next.Array, obj.Array...)
		case cmdReplace:
			rv.secret = cmd.secret
			rv = next
		case cmdKeep:
		case cmdFail:
			return Object{}, ErrConflict
		}
		return rv, nil
	}

	rv := obj
	switch cmd.cmd {
	case cmdSplice, "":
		for key, val := range next.Map {
			newCmd, err := getValidCmd(key, val)
			if err != nil {
				return Object{}, err
			}

			existing, found := obj.Map[newCmd.final]
			if !found {
				// Merging with no conflicts.
				v, err := val.resolveCommands(newCmd.secret)
				if err != nil {
					return Object{}, err
				}
				obj.Map[newCmd.final] = v
				continue
			}

			if existing.Kind() == val.Kind() {
				v, err := existing.merge(newCmd, val)
				if err != nil {
					return Object{}, err
				}
				obj.Map[newCmd.final] = v
				continue
			}

			switch newCmd.cmd {
			case cmdSplice, cmdReplace, "":
				v, err := val.resolveCommands(newCmd.secret)
				if err != nil {
					return Object{}, err
				}
				obj.Map[newCmd.final] = v
			case cmdKeep:
				obj.Map[newCmd.final] = existing
			case cmdFail:
				return Object{}, ErrConflict
			}
		}
	case cmdReplace:
		v, err := next.resolveCommands(false)
		if err != nil {
			return Object{}, err
		}
		rv = v
	case cmdKeep:
	case cmdFail:
		return Object{}, ErrConflict
	}

	rv.secret = cmd.secret
	return rv, nil
}

// getValidCmd gets the command from the key string and validates it is supported.
func getValidCmd(key string, obj Object) (command, error) {
	cmd, err := getCmd(key)
	if err != nil {
		return command{}, err
	}

	list := map[int][]string{
		Map:   {"", cmdFail, cmdKeep, cmdReplace, cmdSplice},
		Array: {"", cmdFail, cmdKeep, cmdReplace, cmdAppend, cmdPrepend},
		Value: {"", cmdFail, cmdKeep, cmdReplace},
	}

	opts, found := list[obj.Kind()]
	if found {
		for _, opt := range opts {
			if cmd.cmd == opt {
				return cmd, nil
			}
		}
	}

	return command{}, ErrInvalidCommand
}