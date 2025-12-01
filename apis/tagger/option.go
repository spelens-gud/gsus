package tagger

const (
	TypeCamelCase = 1
	TypeSnakeCase = 0
)

type TagOption struct {
	Tag        string                                                    `json:"tag" gorm:"column:tag" bson:"tag" form:"tag"`
	Type       int                                                       `json:"type" gorm:"column:type" bson:"type" form:"type"`
	Cover      bool                                                      `json:"cover" gorm:"column:cover" bson:"cover" form:"cover"` //cover old tag
	Edit       bool                                                      `json:"edit" gorm:"column:edit" bson:"edit" form:"edit"`     //edit tag
	AppendFunc func(structName, fieldName, newTag, oldTag string) string `json:"append_func" form:"append_func" gorm:"column:append_func" bson:"append_func"`
}

func CamelCase(tag string, cover bool) TagOption {
	return TagOption{
		Type:  TypeCamelCase,
		Tag:   tag,
		Cover: cover,
		Edit:  true,
	}
}

func SnakeCase(tag string, cover bool) TagOption {
	return TagOption{
		Type:  TypeSnakeCase,
		Tag:   tag,
		Cover: cover,
		Edit:  true,
	}
}
