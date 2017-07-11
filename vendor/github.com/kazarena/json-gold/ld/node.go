package ld

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Node is the value of a subject, predicate or object
// i.e. a IRI reference, blank node or literal.
type Node interface {
	// GetValue returns the node's value.
	GetValue() string

	// Equal returns true id this node is equal to the given node.
	Equal(n Node) bool
}

// Literal represents a literal value.
type Literal struct {
	Value    string
	Datatype string
	Language string
}

// NewLiteral creates a new instance of Literal.
func NewLiteral(value string, datatype string, language string) *Literal {
	l := &Literal{
		Value:    value,
		Language: language,
	}

	if datatype != "" {
		l.Datatype = datatype
	} else {
		l.Datatype = XSDString
	}

	return l
}

// GetValue returns the node's value.
func (l Literal) GetValue() string {
	return l.Value
}

// Equal returns true id this node is equal to the given node.
func (l Literal) Equal(n Node) bool {
	ol, ok := n.(*Literal)
	if !ok {
		return false
	}

	if l.Value != ol.Value {
		return false
	}

	if l.Language != ol.Language {
		return false
	}

	if l.Datatype != ol.Datatype {
		return false
	}

	return true
}

// IRI represents an IRI value.
type IRI struct {
	Value string
}

// NewIRI creates a new instance of IRI.
func NewIRI(iri string) *IRI {
	i := &IRI{
		Value: iri,
	}

	return i
}

// GetValue returns the node's value.
func (iri IRI) GetValue() string {
	return iri.Value
}

// Equal returns true id this node is equal to the given node.
func (iri IRI) Equal(n Node) bool {
	if oiri, ok := n.(*IRI); ok {
		return iri.Value == oiri.Value
	}

	return false
}

// BlankNode represents a blank node value.
type BlankNode struct {
	Attribute string
}

// NewBlankNode creates a new instance of BlankNode.
func NewBlankNode(attribute string) *BlankNode {
	bn := &BlankNode{
		Attribute: attribute,
	}

	return bn
}

// GetValue returns the node's value.
func (bn BlankNode) GetValue() string {
	return bn.Attribute
}

// Equal returns true id this node is equal to the given node.
func (bn BlankNode) Equal(n Node) bool {
	if obn, ok := n.(*BlankNode); ok {
		return bn.Attribute == obn.Attribute
	}

	return false
}

// IsBlankNode returns true if the given node is a blank node
func IsBlankNode(node Node) bool {
	_, isBlankNode := node.(*BlankNode)
	return isBlankNode
}

// IsIRI returns true if the given node is an IRI node
func IsIRI(node Node) bool {
	_, isIRI := node.(*IRI)
	return isIRI
}

// IsLiteral returns true if the given node is a literal node
func IsLiteral(node Node) bool {
	_, isLiteral := node.(*Literal)
	return isLiteral
}

var patternInteger = regexp.MustCompile("^[\\-+]?[0-9]+$")
var patternDouble = regexp.MustCompile("^(\\+|-)?([0-9]+(\\.[0-9]*)?|\\.[0-9]+)([Ee](\\+|-)?[0-9]+)?$")

// rdfToObject converts an RDF triple object to a JSON-LD object.
func rdfToObject(n Node, useNativeTypes bool) (map[string]interface{}, error) {
	// If value is an an IRI or a blank node identifier, return a new
	// JSON object consisting
	// of a single member @id whose value is set to value.
	if IsIRI(n) || IsBlankNode(n) {
		return map[string]interface{}{
			"@id": n.GetValue(),
		}, nil
	}

	literal := n.(*Literal)

	// convert literal object to JSON-LD
	rval := map[string]interface{}{
		"@value": literal.GetValue(),
	}

	// add language
	if literal.Language != "" {
		rval["@language"] = literal.Language
	} else {
		// add datatype
		datatype := literal.Datatype
		value := literal.Value
		if useNativeTypes {
			// use native datatypes for certain xsd types
			if datatype == XSDString {
				// don't add xsd:string
			} else if datatype == XSDBoolean {
				if value == "true" {
					rval["@value"] = true
				} else if value == "false" {
					rval["@value"] = false
				} else {
					// Else do not replace the value, and add the
					// boolean type in
					rval["@type"] = datatype
				}
			} else if (datatype == XSDInteger && patternInteger.MatchString(value)) /* http://www.w3.org/TR/xmlschema11-2/#integer */ ||
				(datatype == XSDDouble && patternDouble.MatchString(value)) /* http://www.w3.org/TR/xmlschema11-2/#nt-doubleRep */ {
				d, _ := strconv.ParseFloat(value, 64)
				if !math.IsNaN(d) && !math.IsInf(d, 0) {
					if datatype == XSDInteger {
						i := int(d)
						if fmt.Sprintf("%d", i) == value {
							rval["@value"] = i
						}
					} else if datatype == XSDDouble {
						rval["@value"] = d
					} else {
						return nil, NewJsonLdError(ParseError, nil)
					}
				}
			} else {
				// do not add xsd:string type
				rval["@type"] = datatype
			}
		} else if datatype != XSDString {
			rval["@type"] = datatype
		}
	}

	return rval, nil
}

// objectToRDF converts a JSON-LD value object to an RDF literal or a JSON-LD string or
// node object to an RDF resource.
func objectToRDF(item interface{}) Node {
	// convert value object to RDF
	if IsValue(item) {
		itemMap := item.(map[string]interface{})
		value := itemMap["@value"]
		datatype := itemMap["@type"]

		// convert to XSD datatypes as appropriate
		booleanVal, isBool := value.(bool)
		floatVal, isFloat := value.(float64)

		if !isBool && !isFloat {
			// if document was created using a standard JSON decoder from json package
			// we need to be careful with float and integer representations.
			// If the client code sets UseNumber() property of json.Decoder
			// (see https://golang.org/pkg/encoding/json/#Decoder.UseNumber )
			// the logic above for discovering floats and integers will fail
			// because they would be represented as json.Number and not float64.
			// The code below takes care of it so it doesn't matter
			// how the document was decoded from JSON.
			if number, isNumber := value.(json.Number); isNumber {
				var floatErr error
				floatVal, floatErr = number.Float64()
				isFloat = (floatErr == nil)
			}
		}

		isInteger := isFloat && floatVal == float64(int64(floatVal))

		datatypeStr, _ := datatype.(string)
		if isBool || isFloat {
			// convert to XSD datatype
			if isBool {
				if datatype == nil {
					return NewLiteral(strconv.FormatBool(booleanVal), XSDBoolean, "")
				} else {
					return NewLiteral(strconv.FormatBool(booleanVal), datatypeStr, "")
				}
			} else if (isFloat && !isInteger) || XSDDouble == datatypeStr {
				canonicalDouble := GetCanonicalDouble(floatVal)
				if datatype == nil {
					return NewLiteral(canonicalDouble, XSDDouble, "")
				} else {
					return NewLiteral(canonicalDouble, datatypeStr, "")
				}
			} else {
				if datatype == nil {
					return NewLiteral(fmt.Sprintf("%d", int(floatVal)), XSDInteger, "")
				} else {
					return NewLiteral(fmt.Sprintf("%d", int(floatVal)), datatype.(string), "")
				}
			}
		} else if langVal, hasLang := itemMap["@language"]; hasLang {
			if datatype == nil {
				return NewLiteral(value.(string), RDFLangString, langVal.(string))
			} else {
				return NewLiteral(value.(string), datatype.(string), langVal.(string))
			}
		} else {
			if datatype == nil {
				return NewLiteral(value.(string), XSDString, "")
			} else {
				return NewLiteral(value.(string), datatype.(string), "")
			}
		}
	} else {
		// convert string/node object to RDF
		id := ""
		if itemMap, isMap := item.(map[string]interface{}); isMap {
			id = itemMap["@id"].(string)
			if IsRelativeIri(id) {
				return nil
			}
		} else {
			id = item.(string)
		}
		if strings.Index(id, "_:") == 0 {
			// NOTE: once again no need to rename existing blank nodes
			return NewBlankNode(id)
		} else {
			return NewIRI(id)
		}
	}
}
