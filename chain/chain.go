package chain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

// Chain contains sentence fragments that can be recombined.
type Chain struct {
	Nodes     []*Node
	PrefixLen int `json:"prefix_len"`
}

// Build creates a new chain from newline delimited text.
func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	p := prefix(make([]string, c.PrefixLen))
	var s string
	for {
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		key := p.key()
		c.appendFragment(key, s)
		p.shift(s)
	}
}

// Generate creates multiple paragraphs.
func (c *Chain) Generate() string {
	n := paragraphCount()
	paragraphs := make([]string, n, n)
	for i := 0; i < n; i++ {
		paragraphs = append(paragraphs, c.GenerateParagraph())
	}
	return strings.TrimLeft(strings.Join(paragraphs, "\n\n"), "\n ")
}

// GenerateParagraph creates a single paragraph.
func (c *Chain) GenerateParagraph() string {
	p := make(prefix, c.PrefixLen)
	n := wordCount()

	var words []string
	for {
		choices := c.fragmentsFor(p.key())
		if choices == nil {
			break
		}
		current := choices[rand.Intn(len(choices))]
		words = append(words, current)
		p.shift(current)
		if len(words) > n && (completesSentence(current) || len(words) > n+500) {
			break
		}
	}
	return strings.Join(words, " ")
}

// ToFile marshalls a chain to a file in JSON format.
func (c *Chain) ToFile(path string) error {
	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0644)
}

// findNode sifts through all the nodes in a chain, looking for a particular key.
// This is used primarily in two contexts: building a chain, and generating output.
// When building a chain, each new word that is encountered must be appended as a
// possible choice following a particular key (one or more words).
// When generating new output, each chosen word becomes the new anchoring point which
// decides what other words may follow.
func (c *Chain) findNode(key string) *Node {
	for _, n := range c.Nodes {
		if n.Key == key {
			return n
		}
	}
	return nil
}

// fragments returns the possible words that may follow a key.
func (c *Chain) fragmentsFor(key string) []string {
	node := c.findNode(key)
	if node == nil {
		return nil
	}
	return node.Fragments
}

// appendFragment adds a word to the list of possible choices following key.
func (c *Chain) appendFragment(key, fragment string) {
	node := c.findNode(key)
	if node == nil {
		node = new(Node)
		node.Key = key
		c.Nodes = append(c.Nodes, node)
	}
	node.Fragments = append(node.Fragments, fragment)
}

// NewChain creates an empty chain.
// The prefix of each node will be prefixLen words long.
func NewChain(prefixLen int) *Chain {
	return &Chain{[]*Node{}, prefixLen}
}

// FromFile unmarshalls a stored JSON chain.
func FromFile(path string) (*Chain, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := new(Chain)

	if err := json.Unmarshal(bytes, c); err != nil {
		return nil, err
	}
	return c, nil
}

// wordCount represents paragraph sizes of unsurprising length.
// The paragraphs are intended to be used in fake comments, and
// paragraphs that are somewhere in this range don't seem particularly
// short, nor particularly long.
func wordCount() int {
	numbers := []int{13, 21, 34, 55, 89, 144}
	return numbers[rand.Intn(len(numbers))]
}

// paragraphCount suggests a reasonable comment size.
func paragraphCount() int {
	numbers := []int{1, 2, 2, 3, 3, 3, 4, 4, 5}
	return numbers[rand.Intn(len(numbers))]
}

// completesSentence determines whether or not a fragment ends with punctuation.
func completesSentence(s string) bool {
	return strings.LastIndexAny(s, "?!.)") == len(s)-1
}
