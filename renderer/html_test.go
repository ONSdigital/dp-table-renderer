package renderer

import (
	"testing"

	"github.com/ONSdigital/dp-table-renderer/models"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"

	"bytes"

	"github.com/ONSdigital/dp-table-renderer/testdata"
	"golang.org/x/net/html/atom"
)

func TestRenderHTML(t *testing.T) {
	t.Parallel()
	Convey("Successfully render an html table", t, func() {
		reader := bytes.NewReader(testdata.LoadExampleRequest(t))
		renderRequest, err := models.CreateRenderRequest(reader)
		if err != nil {
			t.Fatal(err)
		}

		response, err := RenderHTML(renderRequest)
		So(err, ShouldBeNil)
		nodes, err := html.ParseFragment(bytes.NewReader(response), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)
		node := nodes[0]
		So(node.DataAtom, ShouldEqual, atom.Div)
		So(getAttribute(node, "class"), ShouldEqual, "table-renderer")
		table := findNode(node, atom.Table)
		So(table, ShouldNotBeNil)
		rows := findNodes(table, atom.Tr)
		So(len(rows), ShouldEqual, 14)
	})
}

func TestStartTable(t *testing.T) {
	t.Parallel()
	Convey("A table should be described by its subtitle", t, func() {
		request := models.RenderRequest{Filename: "filename", Title: "Heading", Subtitle: "Subtitle"}
		var buf bytes.Buffer

		startTable(&request, &buf)
		nodes, err := html.ParseFragment(bytes.NewReader(buf.Bytes()), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)

		table := nodes[0]
		So(table.DataAtom, ShouldEqual, atom.Table)
		So(getAttribute(table, "id"), ShouldEqual, "table_filename")
		So(getAttribute(table, "aria-describedby"), ShouldEqual, "table_filename_description")
		caption := findNode(table, atom.Caption)
		So(caption, ShouldNotBeNil)
		So(caption.FirstChild.Data, ShouldEqual, "Heading")
		span := findNode(caption, atom.Span)
		So(span, ShouldNotBeNil)
		So(span.FirstChild.Data, ShouldEqual, "Subtitle")
		So(getAttribute(span, "id"), ShouldEqual, "table_filename_description")
	})

	Convey("A table without subtitle should not have aria-describedby", t, func() {
		request := models.RenderRequest{Filename: "filename", Title: "Heading"}
		var buf bytes.Buffer

		startTable(&request, &buf)
		nodes, err := html.ParseFragment(bytes.NewReader(buf.Bytes()), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)

		table := nodes[0]
		So(getAttribute(table, "aria-describedby"), ShouldEqual, "")
		caption := findNode(table, atom.Caption)
		So(caption, ShouldNotBeNil)
		So(findNode(caption, atom.Span), ShouldBeNil)
	})

	Convey("A table without title or subtitle should not have a caption", t, func() {
		request := models.RenderRequest{Filename: "filename"}
		var buf bytes.Buffer

		startTable(&request, &buf)
		nodes, err := html.ParseFragment(bytes.NewReader(buf.Bytes()), &html.Node{
			Type:     html.ElementNode,
			Data:     "body",
			DataAtom: atom.Body,
		})
		So(err, ShouldBeNil)
		So(len(nodes), ShouldBeGreaterThanOrEqualTo, 1)

		table := nodes[0]
		So(getAttribute(table, "aria-describedby"), ShouldEqual, "")
		So(findNode(table, atom.Caption), ShouldBeNil)
	})
}

func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// depth-first search for the first node of the given type
func findNode(n *html.Node, a atom.Atom) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == a {
			return c
		}
		gc := findNode(c, a)
		if gc != nil {
			return gc
		}
	}
	return nil
}

// returns all child nodes of the given type
func findNodes(n *html.Node, a atom.Atom) []*html.Node {
	var result []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == a {
			result = append(result, c)
		}
		result = append(result, findNodes(c, a)...)
	}
	return result
}
