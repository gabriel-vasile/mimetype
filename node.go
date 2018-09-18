package mimetype

import "fmt"

type (
	// Node represents a node in the matchers tree structure.
	// It holds the mime type, the extension and the function to check whether
	// a byte slice has the mime type
	Node struct {
		mime       string
		extension  string
		matchFunc  matchFunc
		exhaustive bool
		children   []*Node
	}
	matchFunc func([]byte) bool
)

// NewNode creates a new Node
func NewNode(mime, extension string, matchFunc matchFunc, children ...*Node) *Node {
	return &Node{
		mime:      mime,
		extension: extension,
		matchFunc: matchFunc,
		children:  children,
	}
}

// Mime returns the mime type associated with the node
func (n *Node) Mime() string { return n.mime }

// Extension returns the file extension associated with the node
func (n *Node) Extension() string { return n.extension }

// Append adds a new node to the matchers tree
// When a node's matching function passes the check, the node's children are
// also checked in order to find a more accurate mime type for the input
func (n *Node) Append(cs ...*Node) { n.children = append(n.children, cs...) }

// match does a depth-first search on the matchers tree
// it returns the deepest successful matcher for which all the children fail
func (n *Node) match(in []byte, deepestMatch *Node) *Node {
	for _, c := range n.children {
		if c.matchFunc(in) {
			return c.match(in, c)
		}
	}

	return deepestMatch
}

// Tree returns a string representation of the matchers tree
func (n *Node) Tree() string {
	var printTree func(*Node, int) string
	printTree = func(n *Node, level int) string {
		offset := ""
		i := 0
		for i < level {
			offset += "|\t"
			i++
		}
		if len(n.children) > 0 {
			offset += "+"
		}
		out := fmt.Sprintf("%s%s \n", offset, n.Mime())
		for _, c := range n.children {
			out += printTree(c, level+1)
		}

		return out
	}

	return printTree(n, 0)
}
