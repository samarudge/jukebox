package mtemplate

type elemlist struct {
	elements []interface{}
}

func NewElemlist() *elemlist {
	newElemList := new(elemlist)
	newElemList.elements = make([]interface{}, 0)
	return newElemList
}

func (this *elemlist) Push(item interface{}) {
	this.elements = append(this.elements, item)
}

func (this elemlist) At(index int) interface{} {
	var foundItem interface{} = nil
	
	if index < len(this.elements) {
		foundItem = this.elements[index]
	}
	
	return foundItem
}

func (this elemlist) Len() int {
	return len(this.elements)
}