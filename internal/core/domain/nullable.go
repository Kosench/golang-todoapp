package domain

// поле в принципе не было передано
// поле передано со значением
// поле передано с щначением null(удаление)
type Nullable[T any] struct {
	Value *T
	Set   bool
}
