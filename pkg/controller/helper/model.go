package helper

import (
	"encoding/json"
	"strings"
	"textract-mongo/pkg/model"
	"textract-mongo/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

type RawKeySet struct {
	Key      string
	Child    []string
	ValueIds []string
}

func MapFormType(blocks []model.Block) []model.KeyValueSet {
	words := map[string]string{}
	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypeWord) && block.Id != nil && block.Text != nil && len(*block.Id) > 0 && len(*block.Text) > 0 {
			words[*block.Id] = *block.Text
		}
	}

	keyValueSet := []model.KeyValueSet{}

	keySet := map[string]RawKeySet{}
	valueSet := map[string][]string{}
	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypeKeyValueSet) && block.Id != nil {
			if utils.Contains(block.EntityTypes, string(types.EntityTypeKey)) {
				keyChild := []string{}
				keyValueIds := []string{}
				for _, relation := range block.Relationships {
					if relation.Type == string(types.RelationshipTypeChild) {
						for _, child := range relation.Ids {
							if val, ok := words[child]; ok {
								keyChild = append(keyChild, val)
							}
						}
					} else if relation.Type == string(types.RelationshipTypeValue) {
						keyValueIds = append(keyValueIds, relation.Ids...)
					}
				}
				keySet[*block.Id] = RawKeySet{
					Key:      strings.Join(keyChild, " "),
					Child:    keyChild,
					ValueIds: keyValueIds,
				}
			} else if utils.Contains(block.EntityTypes, string(types.EntityTypeValue)) {
				valueChild := []string{}
				for _, relation := range block.Relationships {
					if relation.Type == string(types.RelationshipTypeChild) {
						for _, child := range relation.Ids {
							if val, ok := words[child]; ok {
								valueChild = append(valueChild, val)
							}
						}
					}
				}
				valueSet[*block.Id] = valueChild
			}
		}
	}

	for _, key := range keySet {
		kvSet := model.KeyValueSet{
			Key: key.Key,
		}
		vals := []string{}
		for _, val := range key.ValueIds {
			if val, ok := valueSet[val]; ok {
				vals = append(vals, val...)
			}
		}
		kvSet.Value = strings.Join(vals, " ")
		keyValueSet = append(keyValueSet, kvSet)
	}

	return keyValueSet
}

func OutputBlockToModel(blocks []types.Block) []model.Block {
	blockModels := []model.Block{}

	for _, v := range blocks {
		entType := []string{}
		for _, ent := range v.EntityTypes {
			entType = append(entType, string(ent))
		}

		relations := []model.Relationship{}
		for _, r := range v.Relationships {
			relations = append(relations, model.Relationship{
				Ids:  r.Ids,
				Type: string(r.Type),
			})
		}
		var query *model.Query
		if v.Query != nil {
			query = &model.Query{
				Text:  v.Query.Text,
				Alias: v.Query.Alias,
				Pages: v.Query.Pages,
			}
		}
		var geometry *model.Geometry
		if v.Geometry != nil {
			geometry = &model.Geometry{}
			if v.Geometry.BoundingBox != nil {
				geometry.BoundingBox = &model.BoundingBox{
					Height: v.Geometry.BoundingBox.Height,
					Left:   v.Geometry.BoundingBox.Left,
					Width:  v.Geometry.BoundingBox.Width,
					Top:    v.Geometry.BoundingBox.Top,
				}
			}
			if v.Geometry.Polygon != nil {
				points := []model.Point{}
				for _, p := range v.Geometry.Polygon {
					points = append(points, model.Point{
						X: p.X,
						Y: p.Y,
					})
				}
				geometry.Polygon = points
			}
		}

		block := model.Block{
			Id:              v.Id,
			BlockType:       string(v.BlockType),
			ColumnIndex:     v.ColumnIndex,
			ColumnSpan:      v.ColumnSpan,
			RowIndex:        v.RowIndex,
			RowSpan:         v.RowSpan,
			Confidence:      v.Confidence,
			EntityTypes:     entType,
			Geometry:        geometry,
			Relationships:   relations,
			Page:            v.Page,
			Text:            v.Text,
			SelectionStatus: string(v.SelectionStatus),
			TextType:        string(v.TextType),
			Query:           query,
		}

		blockModels = append(blockModels, block)
	}
	return blockModels
}

func GetResponseOfBedrockConverseOutput(converseOutput *bedrockruntime.ConverseOutput) ([]string, error) {
	outputMap := map[string]any{}

	str, err := json.Marshal(converseOutput)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(str), &outputMap)

	output := map[string]any{}
	if val, ok := outputMap["Output"].(map[string]any); ok {
		output = val
	}
	value := map[string]any{}
	if val, ok := output["Value"].(map[string]any); ok {
		value = val
	}
	content := []any{}
	if val, ok := value["Content"].([]any); ok {
		content = val
	}

	result := []string{}
	for _, v := range content {
		txt := ""
		if val, ok := v.(map[string]any); ok {
			if tx, ok2 := val["Value"].(string); ok2 {
				txt = tx
			}
		}
		if len(txt) > 0 {
			result = append(result, txt)
		}
	}

	return result, nil
}
