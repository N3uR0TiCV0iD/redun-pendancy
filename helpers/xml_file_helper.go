package helpers

import (
	"fmt"
	"os"
	"redun-pendancy/utils"

	"github.com/beevik/etree"
)

type XMLWriter interface {
	Write(document *etree.Document) error
}

type XMLFileHelper struct {
	FilePath string
	Document *etree.Document
}

func NewXMLFile(filePath string) (*XMLFileHelper, error) {
	_, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return &XMLFileHelper{
		FilePath: filePath,
	}, nil
}

func (xmlFile *XMLFileHelper) IsLoaded() bool {
	return xmlFile.Document != nil
}

func (xmlFile *XMLFileHelper) LoadOrSkip() error {
	if xmlFile.IsLoaded() {
		return nil
	}
	return xmlFile.Reload()
}

func (xmlFile *XMLFileHelper) Reload() error {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(xmlFile.FilePath)
	if err != nil {
		return err
	}
	xmlFile.Document = doc
	return nil
}

func (xmlFile *XMLFileHelper) InsertNode(node etree.Token, sibling etree.Token) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	if sibling == nil {
		return fmt.Errorf("sibling is nil")
	}
	if !xmlFile.IsLoaded() {
		return fmt.Errorf("document not loaded")
	}

	parent := sibling.Parent()
	if parent == nil {
		return fmt.Errorf("provided sibling has no parent")
	}

	parent.InsertChildAt(sibling.Index(), node)
	return nil
}

func (xmlFile *XMLFileHelper) AppendNode(node etree.Token, parent *etree.Element) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	if parent == nil {
		return fmt.Errorf("parent is nil")
	}
	if !xmlFile.IsLoaded() {
		return fmt.Errorf("document not loaded")
	}

	parent.AddChild(node)
	return nil
}

func (xmlFile *XMLFileHelper) SwapNodes(from etree.Token, to etree.Token) error {
	if to == nil {
		return fmt.Errorf("to is nil")
	}
	if from == nil {
		return fmt.Errorf("from is nil")
	}
	if !xmlFile.IsLoaded() {
		return fmt.Errorf("document not loaded")
	}
	if from == to {
		return nil
	}

	toParent := to.Parent()
	if toParent == nil {
		return fmt.Errorf("to node has no parent")
	}

	fromParent := from.Parent()
	if fromParent == nil {
		return fmt.Errorf("from node has no parent")
	}

	toIndex := to.Index()
	fromIndex := from.Index()
	toParent.RemoveChildAt(toIndex)
	fromParent.RemoveChildAt(fromIndex)

	toParent.InsertChildAt(toIndex, from)
	fromParent.InsertChildAt(fromIndex, to)
	return nil
}

func (xmlFile *XMLFileHelper) RemoveNode(node etree.Token, removeCharData bool, removeParentIfEmpty bool) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	if !xmlFile.IsLoaded() {
		return fmt.Errorf("document not loaded")
	}

	parent := node.Parent()
	if parent == nil {
		return fmt.Errorf("node has no parent")
	}

	index := node.Index()
	parent.RemoveChild(node)
	if !removeCharData {
		return nil
	}

	//Remove adjacent CharData (next sibling if "index == 0", else previous)
	siblingIndex := utils.TernarySelect(index == 0, 0, index-1)
	sibling := parent.Child[siblingIndex]
	_, isCharData := sibling.(*etree.CharData)
	if isCharData {
		parent.RemoveChildAt(siblingIndex)
	}

	if removeParentIfEmpty && xmlFile.IsNodeShallow(parent) {
		xmlFile.RemoveNode(parent, removeCharData, removeParentIfEmpty)
	}
	return nil
}

func (xmlFile *XMLFileHelper) IsNodeShallow(node *etree.Element) bool {
	childCount := len(node.Child)
	if childCount >= 2 {
		return false
	}

	if childCount == 0 {
		return true
	}

	//NOTE: childCount would be 1 here!
	hasChildElements := xmlFile.HasChildElements(node)
	return !hasChildElements
}

func (xmlFile *XMLFileHelper) HasChildElements(node *etree.Element) bool {
	for _, child := range node.Child {
		_, isElement := child.(*etree.Element)
		if isElement {
			return true
		}
	}
	return false
}

func (xmlFile *XMLFileHelper) Commit(xmlWriter XMLWriter) error {
	if xmlFile.Document == nil {
		return fmt.Errorf("document not loaded")
	}

	document := xmlFile.Document
	if xmlWriter == nil {
		err := document.WriteToFile(xmlFile.FilePath)
		return err
	}

	err := xmlWriter.Write(document)
	return err
}
