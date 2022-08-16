// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/schmidtw/goschtalt/pkg/meta"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		description string
		in          string
		expected    meta.Object
		expectedErr error
	}{
		{
			description: "A test of empty.",
			expected:    meta.Object{},
		}, {
			description: "A small test.",
			in: `---
a:
  b:
    c: '123'
d:
  e:
    - fog
    - dog`,
			expected: meta.Object{
				Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 1, Col: 1}},
				Map: map[string]meta.Object{
					"a": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 2, Col: 1}},
						Map: map[string]meta.Object{
							"b": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 3, Col: 3}},
								Map: map[string]meta.Object{
									"c": meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 4, Col: 8}},
										Value:   "123",
									},
								},
							},
						},
					},
					"d": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 5, Col: 1}},
						Map: map[string]meta.Object{
							"e": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 6, Col: 3}},
								Array: []meta.Object{
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 7, Col: 7}},
										Value:   "fog",
									},
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 8, Col: 7}},
										Value:   "dog",
									},
								},
							},
						},
					},
				},
			},
		}, {
			description: "An anchor and merge ... which fails gracefully.",
			in: `---
a:
  b: &foo
    c: '123'
  z: &rat
    y: cat
d:
  <<: *foo
  e: &bar
    - fog
    - dog
  f: &car
    - red
    - blue
g:
  <<: [ *foo, *rat ]
h:
  [*bar, *car]
`,
			expected: meta.Object{
				Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 1, Col: 1}},
				Map: map[string]meta.Object{
					"a": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 2, Col: 1}},
						Map: map[string]meta.Object{
							"b": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 3, Col: 3}},
								Map: map[string]meta.Object{
									"c": meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 4, Col: 8}},
										Value:   "123",
									},
								},
							},
							"z": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 5, Col: 3}},
								Map: map[string]meta.Object{
									"y": meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 6, Col: 8}},
										Value:   "cat",
									},
								},
							},
						},
					},
					"d": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 7, Col: 1}},
						Map: map[string]meta.Object{
							"c": meta.Object{
								Origins: []meta.Origin{},
								Value:   "123",
							},
							"e": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 9, Col: 3}},
								Array: []meta.Object{
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 10, Col: 7}},
										Value:   "fog",
									},
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 11, Col: 7}},
										Value:   "dog",
									},
								},
							},
							"f": meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 12, Col: 3}},
								Array: []meta.Object{
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 13, Col: 7}},
										Value:   "red",
									},
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 14, Col: 7}},
										Value:   "blue",
									},
								},
							},
						},
					},
					"g": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 15, Col: 1}},
						Map: map[string]meta.Object{
							"c": meta.Object{
								Origins: []meta.Origin{},
								Value:   "123",
							},
							"y": meta.Object{
								Origins: []meta.Origin{},
								Value:   "cat",
							},
						},
					},
					"h": meta.Object{
						Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 17, Col: 1}},
						Array: []meta.Object{
							meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 18, Col: 4}},
								Array: []meta.Object{
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 10, Col: 7}},
										Value:   "fog",
									},
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 11, Col: 7}},
										Value:   "dog",
									},
								},
							},
							meta.Object{
								Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 18, Col: 10}},
								Array: []meta.Object{
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 13, Col: 7}},
										Value:   "red",
									},
									meta.Object{
										Origins: []meta.Origin{meta.Origin{File: "file.yml", Line: 14, Col: 7}},
										Value:   "blue",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			var d Decoder
			var got meta.Object
			err := d.Decode("file.yml", []byte(tc.in), &got)

			if tc.expectedErr == nil {
				assert.NoError(err)
				assert.Empty(cmp.Diff(tc.expected, got, cmpopts.IgnoreUnexported(meta.Object{})))
			}
		})
	}
}