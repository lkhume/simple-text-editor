package crdt

import (
	"fmt"
	"sort"
	"sync"
)

// Identifier uniquely identifies an element.
type Identifier struct {
	Site    string `json:"site"`
	Counter int    `json:"counter"`
}

// Compare compares this Identifier with another.
// It returns:
// - -1 if this identifier should come before the other,
// - 0 if they are equal,
// - 1 if it should come after the other
// The comparison is done first by the Counter, and if those are equal, by the Site.
func (id Identifier) Compare(other Identifier) int {
	// Compare the counters first.
	if id.Counter < other.Counter {
		return -1
	} else if id.Counter > other.Counter {
		return 1
	}
	// If the counters are equal, compare the site strings.
	if id.Site < other.Site {
		return -1
	} else if id.Site > other.Site {
		return 1
	}
	return 0
}

// Element represents one character in the collaborative document.
// It contains:
//   - ID: a unique identifier used to determine the element's position
//   - Char: the actual character (stored as a rune)
//   - Deleted: a flag to mark the element as logically deleted (a tombstone), so that we can maintain history and ordering
type Element struct {
	ID      Identifier `json:"id"`
	Char    rune       `json:"char"`
	Deleted bool       `json:"deleted"`
}

// Document represents the entire CRDT document.
type Document struct {
	Elements []Element
	mu       sync.RWMutex
}

// NewDocument creates a new, empty Document
func NewDocument() *Document {
	return &Document{Elements: []Element{}}
}

func (d *Document) Insert(elem Element, pos int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if pos < 0 || pos > len(d.Elements) {
		return fmt.Errorf("invalid position for insertion")
	}

	d.Elements = append(d.Elements, Element{})
	copy(d.Elements[pos+1:], d.Elements[pos:])
	d.Elements[pos] = elem

	return nil
}

func (d *Document) Delete(pos int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if pos < 0 || pos >= len(d.Elements) {
		return fmt.Errorf("invalid position for deletion")
	}

	d.Elements[pos].Deleted = true
	return nil
}

func (d *Document) Merge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	sort.SliceStable(d.Elements, func(i, j int) bool {
		return d.Elements[i].ID.Compare(d.Elements[j].ID) < 0
	})
}

func (d *Document) ToString() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var runes []rune
	for _, elem := range d.Elements {
		// If the element is not deleted, add its character to the slice.
		if !elem.Deleted {
			runes = append(runes, elem.Char)
		}
	}
	// Convert the slice of runes to a string and return it
	return string(runes)
}
