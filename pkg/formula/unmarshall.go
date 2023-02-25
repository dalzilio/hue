// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// ----------------------------------------------------------------------

// A Decoder represents an XML parser reading a particular input stream that
// should be a valid formula file. It embeds an xml.Decoder for parsing a XML
// stream passed as an io.Reader. Like with xml.Decoder, the parser assumes that
// its input is encoded in UTF-8.
type Decoder struct {
	*xml.Decoder
}

// NewDecoder creates a new XML parser reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{xml.NewDecoder(r)}
}

// ----------------------------------------------------------------------

// PropertySet is the type of set of formulas defined in a MCC XML property
// file. We should have a list of 16 properties included inside a `property-set`
// XML element.
type PropertySet struct {
	XMLName xml.Name   `xml:"property-set"`
	List    []Property `xml:"property"`
}

// Property is one of EF phi or AG phi, with phi a state formula
type Property struct {
	Id string `xml:"id"`
	EF RawXML `xml:"formula>exists-path>finally,omitempty"`
	AG RawXML `xml:"formula>all-paths>globally,omitempty"`
}

// StateFormula is a Boolean combination of atomic properties and integer
// expressions that we should parse explicitly.
type RawXML struct {
	InnerXML []byte `xml:",innerxml"`
}

type Places struct {
	PlList []string `xml:"place"`
}

type Transitions struct {
	TrList []string `xml:"transition"`
}

type Constant struct {
	Value []byte `xml:",innerxml"`
}

// ----------------------------------------------------------------------

// parseFormula returns the state formula corresponding to the XML content
// found in the provided byte slice.
func parseFormula(b []byte) (Formula, error) {
	if len(b) == 0 {
		return nil, nil
	}
	buff := bytes.NewBuffer(b)
	decoder := xml.NewDecoder(buff)
	return parseElement(decoder)
}

func parseMult(d *xml.Decoder, acc []Formula) ([]Formula, error) {
	res, err := parseElement(d)
	if err != nil {
		return nil, fmt.Errorf(" parsing multiple elements: %s", err)
	}
	if res == nil {
		return acc, nil
	}
	return parseMult(d, append(acc, res))
}

func parseInner(decoder *xml.Decoder) (Formula, error) {
	res, err := parseElement(decoder)
	skipUntilEnd(decoder)
	return res, err
}

func skipUntilEnd(d *xml.Decoder) error {
	t, _ := d.Token()
	if t == nil {
		return fmt.Errorf("expecting an xml.EndElement")
	}
	switch t.(type) {
	case xml.CharData:
		skipUntilEnd(d)
	case xml.EndElement:
		return nil
	}
	return fmt.Errorf("expecting an xml.EndElement, found (%T) %v", t, t)
}

func parseElement(decoder *xml.Decoder) (Formula, error) {
	t, _ := decoder.Token()
	if t == nil {
		return nil, nil
	}

	switch se := t.(type) {
	case xml.CharData:
		return parseElement(decoder)
	case xml.EndElement:
		return nil, nil
	case xml.StartElement:
		switch se.Name.Local {
		case "negation":
			res, err := parseInner(decoder)
			if err != nil {
				return nil, err
			}
			return Negation{res}, nil
		case "disjunction":
			ee, err := parseMult(decoder, nil)
			if err != nil {
				return nil, err
			}
			// if len(ee) < 2 {
			// 	return nil, errors.New("malformed Formula: disjunction expect at least two sub-formulas")
			// }
			return Disjunction(ee), nil
		case "conjunction":
			ee, err := parseMult(decoder, nil)
			if err != nil {
				return nil, err
			}
			// if len(ee) < 2 {
			// 	return nil, errors.New("malformed Formula: conjunction expect at least two sub-formulas")
			// }
			return Conjunction(ee), nil
		case "integer-sum":
			ee, err := parseMult(decoder, nil)
			if err != nil {
				return nil, err
			}
			// if len(ee) < 2 {
			// 	return nil, errors.New("malformed Formula: integer-sum expect at least two sub-formulas")
			// }
			return IntegerSum(ee), nil
		case "integer-difference":
			ee, err := parseMult(decoder, nil)
			if err != nil {
				return nil, err
			}
			// if len(ee) < 2 {
			// 	return nil, errors.New("malformed Formula: integer-sum expect at least two sub-formulas")
			// }
			return IntegerDifference(ee), nil
		case "integer-le":
			ee, err := parseMult(decoder, nil)
			if err != nil {
				return nil, err
			}
			if len(ee) != 2 {
				return nil, errors.New("malformed Formula: integer-le does not have two operands")
			}
			return IntegerLe{
				Left:  ee[0],
				Right: ee[1],
			}, nil
		case "integer-constant":
			var cst Constant
			err := decoder.DecodeElement(&cst, &se)
			if err != nil {
				return nil, err
			}
			v, err := strconv.Atoi(string(cst.Value))
			if err != nil {
				return nil, err
			}
			return IntegerConstant(v), nil
		case "is-fireable":
			var tr Transitions
			err := decoder.DecodeElement(&tr, &se)
			if err != nil {
				return nil, err
			}
			sort.Strings(tr.TrList)
			return IsFireable(tr.TrList), nil
		case "tokens-count":
			var pl Places
			err := decoder.DecodeElement(&pl, &se)
			if err != nil {
				return nil, err
			}
			sort.Strings(pl.PlList)
			return TokensCount(pl.PlList), nil
		default:
			return nil, errors.New("malformed XML: unexpected Token <" + se.Name.Local + ">")
		}
	}
	return nil, errors.New("malformed XML: unexpected Token in parseFormula")
}

// ----------------------------------------------------------------------

func (d *Decoder) Build() ([]Query, error) {
	decoder := d.Decoder
	var props PropertySet
	err := decoder.Decode(&props)
	if err != nil {
		return nil, fmt.Errorf(" decoding XML input: %s", err)
	}
	res := make([]Query, len(props.List))

	for k, p := range props.List {
		if len(p.EF.InnerXML) != 0 {
			ff, err := parseFormula(p.EF.InnerXML)
			if err != nil {
				return nil, fmt.Errorf(" decoding XML input in EF query %d : %s", k, err)
			}
			res[k] = Query{ID: p.Id, IsEF: true, Original: ff, Formula: Simplify(ff)}
		} else {
			ff, err := parseFormula(p.AG.InnerXML)
			if err != nil {
				return nil, fmt.Errorf(" decoding XML input in AG query %d : %s", k, err)
			}
			res[k] = Query{ID: p.Id, IsEF: false, Original: ff, Formula: Simplify(ff)}
		}
	}

	return res, nil
}
