package transform

type Template struct {
	params         *Parameters
	Transformation *Transformation
}

func NewTemplate(params *Parameters, options TransformationOptions) *Template {
	return &Template{
		params:         params,
		Transformation: NewRandomTransformationConfig(params, options),
	}
}
