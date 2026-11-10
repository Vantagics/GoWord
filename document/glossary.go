package document

// BuildingBlock is one entry in a glossary document (word/glossary/document.xml).
type BuildingBlock struct {
	Name  string
	Style string
	Text  string
}

// AddBuildingBlock appends a glossary building block. The next save writes
// word/glossary/document.xml and the glossary relationship.
func (d *Document) AddBuildingBlock(name, text string) *BuildingBlock {
	if d == nil {
		return nil
	}
	bb := &BuildingBlock{Name: name, Text: text}
	d.buildingBlocks = append(d.buildingBlocks, bb)
	return bb
}

// BuildingBlocks returns glossary entries in document order.
func (d *Document) BuildingBlocks() []*BuildingBlock {
	if d == nil {
		return nil
	}
	return d.buildingBlocks
}
