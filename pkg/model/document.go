package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Document struct {
	ID               primitive.ObjectID `bson:"_id"`
	Name             string             `bson:"name"`
	Blocks           []Block            `bson:"blocks"`
	DocumentMetadata *DocumentMetadata  `bson:"documentMetadata"`
	MappedKeyValue   []KeyValueSet      `bson:"mappedKeyValues"`
	ExtractType      string             `bson:"extractType"`
	Summary          string             `bson:"summary"`
	BedrockResponse  any                `bson:"bedrockResponse"`
}

type DocumentMetadata struct {
	Pages *int32 `bson:"pages"`
}

type KeyValueSet struct {
	Key   string
	Value string
}

type Block struct {
	Id              *string        `bson:"Id"`
	BlockType       string         `bson:"blockType"`
	ColumnIndex     *int32         `bson:"columnIndex"`
	ColumnSpan      *int32         `bson:"columnSpan"`
	RowIndex        *int32         `bson:"rowIndex"`
	RowSpan         *int32         `bson:"rowSpan"`
	Confidence      *float32       `bson:"confidence"`
	EntityTypes     []string       `bson:"entityTypes"`
	Geometry        *Geometry      `bson:"geometry"`
	Page            *int32         `bson:"page"`
	Query           *Query         `bson:"query"`
	Relationships   []Relationship `bson:"relationships"`
	SelectionStatus string         `bson:"selectionStatus"`
	Text            *string        `bson:"text"`
	TextType        string         `bson:"textType"`
}

type Query struct {
	Text  *string
	Alias *string
	Pages []string
}
type Geometry struct {
	BoundingBox *BoundingBox `bson:"BoundingBox"`
	Polygon     []Point      `bson:"Polygon"`
}
type BoundingBox struct {
	Height float32 `bson:"Height"`
	Left   float32 `bson:"Left"`
	Top    float32 `bson:"Top"`
	Width  float32 `bson:"Width"`
}
type Point struct {
	X float32 `bson:"X"`
	Y float32 `bson:"Y"`
}
type Relationship struct {
	Ids  []string `bson:"Ids"`
	Type string   `bson:"Type"`
}
