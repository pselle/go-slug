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

	type file struct {
		// the actual path, should be file path format /dir/subdir/file.extension
		path string
		// is it a directory?
		dir bool
		// should match
		match bool
	}
	paths := []file{
		{
			path:  ".terraform/",
			dir:   true,
			match: true,
		},
		{
			path:  "included.txt",
			match: false,
		},
		{
			path:  ".terraform/foo/bar",
			dir:   false,
			match: true,
		},
		{
			path:  ".terraform/foo/bar/more/directories/so/many",
			dir:   false,
			match: true,
		},
		{
			path:  ".terraform/foo/ignored-subdirectory/",
			dir:   true,
			match: true,
		},
		{
			path:  "baz.txt",
			dir:   false,
			match: true,
		},
		{
			path:  "parent/foo/baz.txt",
			dir:   false,
			match: true,
		},
		// baz.txt is ignored, but a file name including it should not be
		{
			path:  "something/with-baz.txt",
			dir:   false,
			match: false,
		},
		{
			path:  "something/baz.x",
			dir:   false,
			match: false,
		},
		// // ignore sub- terraform.d paths
		// {
		// 	path:  "some-module/terraform.d/x",
		// 	dir:   false,
		// 	match: true,
		// },
		// // but not the root one
		// // {
		// // 	path:  "terraform.d/",
		// // 	dir:   true,
		// // 	match: false,
		// // },
		// {
		// 	// We ignore the directory, but a file of the same name could exist
		// 	path:  "terraform.d",
		// 	dir:   false,
		// 	match: false,
		// },
		// // Getting into * patterns
		// {
		// 	path:  "foo/ignored-doc.md",
		// 	dir:   false,
		// 	match: true,
		// },
		// {
		// 	path:  "foo/otherfile",
		// 	dir:   false,
		// 	match: false,
		// },
	}
	for i, p := range paths {
		match := matchIgnorePattern(p.path, ignorePatterns)
		if match != p.match {
			t.Fatalf("%s at index %d should be %t", p.path, i, p.match)
		}
	}
}
