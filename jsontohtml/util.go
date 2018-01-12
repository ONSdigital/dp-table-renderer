package jsontohtml

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// createNode creates an html Node and sets attributes or adds child nodes according to the type of each value
func createNode(data string, dataAtom atom.Atom, values ...interface{}) *html.Node {
	node := &html.Node{
		Type:     html.ElementNode,
		Data:     data,
		DataAtom: dataAtom,
	}
	for _, value := range values {
		switch v := value.(type) {
		case html.Attribute:
			node.Attr = append(node.Attr, v)
		case *html.Node:
			node.AppendChild(v)
		case []*html.Node:
			for _, c := range v {
				node.AppendChild(c)
			}
		case string:
			node.AppendChild(&html.Node{Type: html.TextNode, Data: v})
		}
	}
	return node
}

// addAttribute adds an attribute to the node
func addAttribute(node *html.Node, key string, val string) {
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: val})
}

// replaceAttribute adds an attribute to the node, replacing any existing attribute with the same name
func replaceAttribute(node *html.Node, key string, val string) {
	var attr []html.Attribute
	for _, a := range node.Attr {
		if a.Key != key {
			attr = append(attr, a)
		}
	}
	node.Attr = append(attr, html.Attribute{Key: key, Val: val})
}

// attr creates a new Attribute
func attr(key string, val string) html.Attribute {
	return html.Attribute{Key: key, Val: val}
}

// text creates a new text node
func text(text string) *html.Node {
	return &html.Node{Type: html.TextNode, Data: text}
}

// getAttribute finds an attribute for the node - returns empty string if not found
func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// findNode is a depth-first search for the first node of the given type
func findNode(n *html.Node, a atom.Atom) *html.Node {
	return findNodeWithAttributes(n, a, nil)
}

// findNodeWithAttributes is a depth-first search for the first node of the given type with the given attributes
func findNodeWithAttributes(n *html.Node, a atom.Atom, attr map[string]string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == a && hasAttributes(c, attr) {
			return c
		}
		gc := findNodeWithAttributes(c, a, attr)
		if gc != nil {
			return gc
		}
	}
	return nil
}

// hasAttributes returns true if the given node has all the attribute values
func hasAttributes(n *html.Node, attr map[string]string) bool {
	for key, value := range attr {
		if getAttribute(n, key) != value {
			return false
		}
	}
	return true
}

// findNodes returns all child nodes of the given type
func findNodes(n *html.Node, a atom.Atom) []*html.Node {
	return findNodesWithAttributes(n, a, nil)
}

// findNodesWithAttributes returns all child nodes of the given type with the given attributes
func findNodesWithAttributes(n *html.Node, a atom.Atom, attr map[string]string) []*html.Node {
	var result []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == a && hasAttributes(c, attr) {
			result = append(result, c)
		}
		result = append(result, findNodesWithAttributes(c, a, attr)...)
	}
	return result
}
