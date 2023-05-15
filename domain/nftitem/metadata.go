package nftitem

type Attribute struct {
	TraitType   string `json:"trait_type" bson:"trait_type"`
	Value       string `json:"value" bson:"value"`
	DisplayType string `json:"display_type,omitempty" bson:"display_type,omitempty"`
}

type RawAttribute struct {
	TraitType   string      `json:"trait_type" bson:"trait_type"`
	Value       interface{} `json:"value" bson:"value"`
	DisplayType string      `json:"display_type,omitempty" bson:"display_type,omitempty"`
}

type Attributes = []Attribute

type PropertyDetail struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type Properties = map[string]interface{}

type PropertyDetails = map[string]PropertyDetail
