package helpers

import (
	"fmt"
	"io"
	"strings"

	"github.com/beevik/etree"
)

type CustomXMLWriter struct {
	Indentation     string
	AddEmptyEOFLine bool

	writer io.Writer
}

func NewCustomXMLWriter(writer io.Writer, indentation string, addEmptyEOFLine bool) *CustomXMLWriter {
	return &CustomXMLWriter{
		Indentation:     indentation,
		AddEmptyEOFLine: addEmptyEOFLine,

		writer: writer,
	}
}

func (xmlWriter *CustomXMLWriter) Write(document *etree.Document) error {
	lastCharData, _, err := xmlWriter.writeChildren(document.Child, 0)
	if err != nil {
		return err
	}

	if !xmlWriter.AddEmptyEOFLine {
		return nil
	}

	err = xmlWriter.writeMissingIndentation(lastCharData, "\n")
	return err
}

func (xmlWriter *CustomXMLWriter) writeChildren(children []etree.Token, level int) (string, bool, error) {
	var err error
	lastCharData := ""
	hasElements := false
	expectedIndentation := xmlWriter.buildIndentation(level)
	for _, child := range children {
		switch child := child.(type) {
		case *etree.Comment:
			_, err = io.WriteString(xmlWriter.writer, "<!--"+child.Data+"-->")
		case *etree.Element:
			err = xmlWriter.writeMissingIndentation(lastCharData, expectedIndentation)
			if err != nil {
				return "", false, err
			}
			err = xmlWriter.writeElement(child, expectedIndentation, level)
			lastCharData = ""
			hasElements = true
		case *etree.CharData:
			lastCharData = child.Data
			_, err = io.WriteString(xmlWriter.writer, lastCharData)
		}
		if err != nil {
			return "", false, err
		}
	}
	return lastCharData, hasElements, nil
}

func (xmlWriter *CustomXMLWriter) buildIndentation(level int) string {
	if level == 0 || xmlWriter.Indentation == "" {
		return ""
	}
	return strings.Repeat(xmlWriter.Indentation, level)
}

func (xmlWriter *CustomXMLWriter) writeMissingIndentation(lastCharData string, expectedIndentation string) error {
	if strings.HasSuffix(lastCharData, expectedIndentation) {
		return nil
	}

	missingIndentation := calculateMissingIndentation(lastCharData, expectedIndentation)
	_, err := io.WriteString(xmlWriter.writer, missingIndentation)
	return err
}

func calculateMissingIndentation(lastCharData string, expectedIndentation string) string {
	// Match characters in reverse until they no longer match or indices are out of bounds
	lastIndex := len(lastCharData) - 1
	expectedIndex := len(expectedIndentation) - 1
	for lastIndex >= 0 && expectedIndex >= 0 && lastCharData[lastIndex] == expectedIndentation[expectedIndex] {
		lastIndex--
		expectedIndex--
	}
	return expectedIndentation[:expectedIndex+1]
}

func (xmlWriter *CustomXMLWriter) writeElement(element *etree.Element, expectedIndentation string, level int) error {
	_, err := io.WriteString(xmlWriter.writer, "<"+element.Tag)
	if err != nil {
		return err
	}

	err = xmlWriter.writeAttributes(element)
	if err != nil {
		return err
	}

	if len(element.Child) == 0 {
		//Self-closing tag
		_, err = io.WriteString(xmlWriter.writer, " />")
		return err
	}

	_, err = io.WriteString(xmlWriter.writer, ">")
	if err != nil {
		return err
	}

	lastCharData, hasElements, err := xmlWriter.writeChildren(element.Child, level+1)
	if err != nil {
		return err
	}

	if hasElements {
		err = xmlWriter.writeMissingIndentation(lastCharData, expectedIndentation)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(xmlWriter.writer, "</"+element.Tag+">")
	return err
}

func (xmlWriter *CustomXMLWriter) writeAttributes(element *etree.Element) error {
	for _, attribute := range element.Attr {
		attributeText := fmt.Sprintf(` %s="%s"`, attribute.Key, attribute.Value)
		_, err := io.WriteString(xmlWriter.writer, attributeText)
		if err != nil {
			return err
		}
	}
	return nil
}
