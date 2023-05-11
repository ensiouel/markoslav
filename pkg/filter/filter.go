package filter

type Operator int

const (
	OperatorEq Operator = iota
	OperatorNotEq
)

type Field struct {
	Name     string
	Value    any
	Operator Operator
}

type options struct {
	fields []Field
}

func NewOptions() Options {
	return &options{}
}

type Options interface {
	Add(name string, value any, operator Operator) Options
	Fields() []Field
}

func (o *options) Add(name string, value any, operator Operator) Options {
	o.fields = append(o.fields, Field{
		Name:     name,
		Value:    value,
		Operator: operator,
	})

	return o
}

func (o *options) Fields() []Field {
	return o.fields
}
