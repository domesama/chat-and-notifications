package stub

import "context"

type Predicate[T any] func(context.Context, T) bool

type Predicates[T any] []Predicate[T]

func NewPredicates[T any](predicate ...Predicate[T]) Predicates[T] {
	return predicate
}

func (p Predicates[T]) IsSatisfied(ctx context.Context, incoming T) bool {
	for _, predicate := range p {
		if !predicate(ctx, incoming) {
			return false
		}
	}
	return true
}
