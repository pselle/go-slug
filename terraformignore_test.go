package slug

import (
	"testing"
)

// Test ignore functions in isolation
func TestTerraformIgnore(t *testing.T) {
	// path to directory without .terraformignore
	// p := parseIgnoreFile("testdata/external-dir")
	// if len(p) != 5 {
	// 	t.Fatal("A directory without .terraformignore should get the default patterns")
	// }

	// load the .terraformignore file's patterns
	ignorePatterns := parseIgnoreFile("testdata/archive-dir")
	// ignorePatterns := []rule{
	// 	{
	// 		pattern:  "foo/*.md",
	// 		excluded: false,
	// 	},
	// }
	type file struct {
		// the actual path, should be file path format /dir/subdir/file.extension
		path string
		// should match
		match bool
	}
	paths := []file{
		{
			path:  ".terraform/",
			match: true,
		},
		{
			path:  "included.txt",
			match: false,
		},
		{
			path:  ".terraform/foo/bar",
			match: true,
		},
		{
			path:  ".terraform/foo/bar/more/directories/so/many",
			match: true,
		},
		{
			path:  ".terraform/foo/ignored-subdirectory/",
			match: true,
		},
		{
			path:  "baz.txt",
			match: true,
		},
		{
			path:  "parent/foo/baz.txt",
			match: true,
		},
		// baz.txt is ignored, but a file name including it should not be
		{
			path:  "something/with-baz.txt",
			match: false,
		},
		{
			path:  "something/baz.x",
			match: false,
		},
		// // ignore sub- terraform.d paths
		// {
		// 	path:  "some-module/terraform.d/x",
		// 	match: true,
		// },
		// // but not the root one
		// // {
		// // 	path:  "terraform.d/",
		// // 	match: false,
		// // },
		{
			// We ignore the directory, but a file of the same name could exist
			path:  "terraform.d",
			match: false,
		},
		// // Getting into * patterns
		{
			path:  "foo/ignored-doc.md",
			match: true,
		},
		{
			path:  "foo/otherfile",
			match: false,
		},
	}
	for i, p := range paths {
		match := matchIgnorePattern(p.path, ignorePatterns)
		if match != p.match {
			t.Fatalf("%s at index %d should be %t", p.path, i, p.match)
		}
	}
}
