// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

// ----------------------------------------------------------------------

// pnml is the type of PNML net. We ignore the graphical information contained
// in the net and only consider the first page
type pnml struct {
	Net Net `xml:"net"`
}

// Net is the type of the net element in a PNML file.
type Net struct {
	Name        string      `xml:"name>text"`
	Page        Page        `xml:"page"`
	Declaration Declaration `xml:"declaration>structure>declarations"`
	// TypeEnvt is an association between a variable name and its type name, found in declaration
	TypeEnvt map[string]string
	// types gives the type (declaration) corresponding to a given constant.
	// This is only used for FENUM and CENUM (for computing predecessors and
	// successors)
	types map[string]*TypeDecl
	// ranges is used to find a suitable finite int range type given the bounds.
	// The idea is that two ranges with the same bounds are isomorphic types.
	ranges map[IntRange]*TypeDecl
	// position tells the position of the constant in its type; used for successor
	position map[string]int
	// order associates a unique Value to every Constant; it is used for
	// encoding Constants into Values
	order map[string]*Value
	// identity associates a string to a constant index. This is only useful for printing
	// debugging information
	identity []string
	// Unique associates a Unique representant for each Value
	Unique map[Value]*Value
	// World associates a type (name) with a list of all its possible values
	World map[string][]*Value
	// vdot is the Value for the dot constant
	vdot *Value
}

// ----------------------------------------------------------------------

// Page is the type of the page element in a PNML file.
type Page struct {
	Places []*Place      `xml:"place"`
	Trans  []*Transition `xml:"transition"`
	Arcs   []*Arc        `xml:"arc"`
}

// Declaration is the type of a PNML net declaration. It contains declarations
// for types and variables used in the net. We also added partitions (and
// partitionelement) from model VehicularWifi
type Declaration struct {
	Sorts      []*TypeDecl      `xml:"namedsort"`
	Vars       []*VarDecl       `xml:"variabledecl"`
	Partitions []*PartitionDecl `xml:"partition"`
}

// ----------------------------------------------------------------------

// TypeDecl is the type of  PNML type declarations. We use a pointer field for
// Dot in order to discriminate to differentiate between default value and field
// initialized. Same with finite int range.
type TypeDecl struct {
	Sort    TYP
	Elem    []string
	ID      string    `xml:"id,attr"`
	CEnum   []Fec     `xml:"cyclicenumeration>feconstant,omitempty"`
	FEnum   []Fec     `xml:"finiteenumeration>feconstant,omitempty"`
	FIntRan *IntRange `xml:"finiteintrange,omitempty"`
	Product []Type    `xml:"productsort>usersort,omitempty"`
	Dot     *struct{} `xml:"dot,omitempty"`
}

// ----------------------------------------------------------------------

// PartitionDecl is the type of  PNML partition declarations.
type PartitionDecl struct {
	ID         string `xml:"id,attr"`
	Type       `xml:"usersort,omitempty"`
	Partitions []Partition `xml:"partitionelement,omitempty"`
}

// Partition list a subset of values of a given (enumeration) type.
type Partition struct {
	ID   string `xml:"id,attr"`
	Elem []Type `xml:"useroperator,omitempty"`
}

// ----------------------------------------------------------------------

// IntRange is the type of PNML int ranges
type IntRange struct {
	Start int `xml:"start,attr"`
	End   int `xml:"end,attr"`
}

// Fec is the type of  PNML enumeration constants.
type Fec struct {
	ID string `xml:"id,attr"`
}

// VarDecl is the type of  PNML variable  declarations.
type VarDecl struct {
	ID   string `xml:"id,attr"`
	Type Type   `xml:"usersort"`
}

// Type is the type of a type declaration.
type Type struct {
	ID string `xml:"declaration,attr"`
}

// ----------------------------------------------------------------------

// Place is the type of a PNML place. It can contain a type and an (optional)
// initial marking.
type Place struct {
	ID             string `xml:"id,attr"`
	Type           Type   `xml:"type>structure>usersort"`
	XML            RawXML `xml:"hlinitialMarking>structure"`
	InitialMarking Expression
}

// RawXML is the type of PNML initial marking expressions, patterns and conditions.
type RawXML struct {
	InnerXML []byte `xml:",innerxml"`
}

// ----------------------------------------------------------------------

// Transition is the type of a PNML transition. It can contain a type and an (optional)
// initial marking.
type Transition struct {
	ID        string `xml:"id,attr"`
	XML       RawXML `xml:"condition>structure"`
	Condition Operation
}

// ----------------------------------------------------------------------

// Arc is the type of edges element in a PNML net.
type Arc struct {
	Source  string `xml:"source,attr"`
	Target  string `xml:"target,attr"`
	XML     RawXML `xml:"hlinscription>structure"`
	Pattern Expression
}

// ----------------------------------------------------------------------

// NumberConstant is used in PNML expressions.
type NumberConstant struct {
	Value int `xml:"value,attr"`
}

// FIRangeConstant is used in PNML expressions.
type FIRangeConstant struct {
	Value int      `xml:"value,attr"`
	Range IntRange `xml:"finiteintrange"`
}

// Variable is used in PNML expressions.
type Variable struct {
	RefVariable string `xml:"refvariable,attr"`
}
