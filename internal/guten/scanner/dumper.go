package scanner

import "fmt"

// Dumper implements a debugging Resultor that prints about each scan event.
type Dumper struct{}

// Close prints DONE
func (d Dumper) Close() error {
	_, err := fmt.Printf("DONE\n")
	return err
}

// Slug prist the first-paragraph slug.
func (d Dumper) Slug(s string) error {
	_, err := fmt.Printf("slug: %q\n", s)
	return err
}

// Meta prints each pre-body key => val pair.
func (d Dumper) Meta(key, val string) error {
	_, err := fmt.Printf("%q => %q\n", key, val)
	return err
}

// Mark prints a structural marker found in the document.
func (d Dumper) Mark(s string) error {
	_, err := fmt.Printf("MARK: %q\n", s)
	return err
}

// Boundary prints the start/end of a structural document area.
func (d Dumper) Boundary(end bool, name string) error {
	kind := "Start"
	if end {
		kind = "End"
	}
	_, err := fmt.Printf("# %s: %q\n", kind, name)
	return err
}

// Data prints a received buffer of line data.
func (d Dumper) Data(buf []byte) error {
	_, err := fmt.Printf("NOPE: %q\n", buf)
	return err
}

var _ Resultor = Dumper{}
