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
		// var geometry *model.Geometry
		// if v.Geometry != nil {
		// 	geometry = &model.Geometry{}
		// 	if v.Geometry.BoundingBox != nil {
		// 		geometry.BoundingBox = &model.BoundingBox{
		// 			Height: v.Geometry.BoundingBox.Height,
		// 			Left:   v.Geometry.BoundingBox.Left,
		// 			Width:  v.Geometry.BoundingBox.Width,
		// 			Top:    v.Geometry.BoundingBox.Top,
		// 		}
		// 	}
		// 	if v.Geometry.Polygon != nil {
		// 		points := []model.Point{}
		// 		for _, p := range v.Geometry.Polygon {
		// 			points = append(points, model.Point{
		// 				X: p.X,
		// 				Y: p.Y,
		// 			})
		// 		}
		// 		geometry.Polygon = points
		// 	}
		// }

		block := model.Block{
			Id:          v.Id,
			BlockType:   string(v.BlockType),
			ColumnIndex: v.ColumnIndex,
			ColumnSpan:  v.ColumnSpan,
			RowIndex:    v.RowIndex,
			RowSpan:     v.RowSpan,
			Confidence:  v.Confidence,
			EntityTypes: entType,
			// Geometry:        geometry,
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

func MapTableType(blocks []model.Block) []model.TableData {
	cells := map[string]model.Block{}
	words := map[string]string{}
	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypeCell) {
			cells[*block.Id] = block
		}
		if block.BlockType == string(types.BlockTypeWord) && block.Id != nil && block.Text != nil && len(*block.Id) > 0 && len(*block.Text) > 0 {
			words[*block.Id] = *block.Text
		}
	}

	// tables := map[string]map[int]map[int]string{}
	// tableStructure := map[string]map[int]string{}

	modelTable := []model.TableData{}

	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypeTable) {
			// fmt.Printf("Table detected: %d\n", block.Id)
			table := map[int]map[int]string{}
			structure := map[int]string{}
			for _, rel := range block.Relationships {
				if rel.Type == string(types.RelationshipTypeChild) {
					for _, child := range rel.Ids {
						if cell, ok := cells[child]; ok {

							// if cell.RowIndex != nil && cell.ColumnIndex != nil {
							// 	fmt.Printf("Cell %d-%d:\n", *cell.RowIndex, *cell.ColumnIndex)
							// }
							line := []string{}
							for _, celRel := range cell.Relationships {
								if celRel.Type == string(types.RelationshipTypeChild) {
									for _, childWord := range celRel.Ids {
										if word, ok := words[childWord]; ok {
											line = append(line, word)
										}
									}
								}
							}
							// fmt.Printf("%s\n", strings.Join(line, " "))

							if cell.RowIndex != nil && cell.ColumnIndex != nil {
								if utils.Contains(cell.EntityTypes, string(types.EntityTypeColumnHeader)) && utils.Contains(block.EntityTypes, string(types.EntityTypeStructuredTable)) {
									structure[int(*cell.ColumnIndex)] = strings.Join(line, " ")
								} else {
									if _, ok := table[int(*cell.RowIndex)]; !ok {
										table[int(*cell.RowIndex)] = map[int]string{}
									}
									table[int(*cell.RowIndex)][int(*cell.ColumnIndex)] = strings.Join(line, " ")
								}
							}
						}
					}
				}
			}
			page := 0
			if block.Page != nil {
				page = int(*block.Page)
			}
			tableType := types.EntityTypeStructuredTable
			if len(structure) < 1 {
				tableType = types.EntityTypeSemiStructuredTable
			}
			modelTable = append(modelTable, model.TableData{
				Data:      table,
				Structure: structure,
				Page:      page,
				TableType: string(tableType),
			})
			// tables[*block.Id] = table
			// tableStructure[*block.Id] = structure
		}
	}

	return modelTable
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
func MapLayoutTextType(blocks []model.Block) []model.PageParagraph {
	blockMap := map[string]model.Block{}
	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypeLine) || block.BlockType == string(types.BlockTypeLayoutText) {
			blockMap[*block.Id] = block
		}
	}

	ignoredText := []string{}
	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypeLayoutTable) {
			for _, rel := range block.Relationships {
				if rel.Type == string(types.RelationshipTypeChild) {
					ignoredText = append(ignoredText, rel.Ids...)
				}
			}
		}
	}

	layoutTextMap := []model.PageParagraph{}
	for _, block := range blocks {
		if block.BlockType == string(types.BlockTypePage) && block.Id != nil {
			paragraph := []string{}
			for _, pageRel := range block.Relationships {
				if pageRel.Type == string(types.RelationshipTypeChild) {
					for _, pageChild := range pageRel.Ids {
						if textLay, ok := blockMap[pageChild]; ok {
							if textLay.BlockType == string(types.BlockTypeLayoutText) {
								for _, texLayRel := range block.Relationships {
									if texLayRel.Type == string(types.RelationshipTypeChild) {
										for _, textLayChild := range texLayRel.Ids {
											if line, ok := blockMap[textLayChild]; ok {
												if line.BlockType == string(types.BlockTypeLine) && !utils.Contains(ignoredText, *line.Id) {
													paragraph = append(paragraph, *line.Text)
													ignoredText = append(ignoredText, *line.Id)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
			if len(paragraph) > 0 {
				page := 0
				if block.Page != nil {
					page = int(*block.Page)
				}
				layoutTextMap = append(layoutTextMap, model.PageParagraph{
					Page:      page,
					Paragraph: strings.Join(paragraph, " "),
				})
			}
		}
	}
	return layoutTextMap
}
