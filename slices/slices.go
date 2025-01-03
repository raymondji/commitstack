package slices

import (
	"slices"
)

func ToMap[K comparable, V any](s []V, keyFunc func(V) K) map[K]V {
	m := map[K]V{}
	for _, v := range s {
		k := keyFunc(v)
		m[k] = v
	}
	return m
}

func Filter[T any](s []T, predicate func(T) bool) []T {
	var out []T
	for _, v := range s {
		if predicate(v) {
			out = append(out, v)
		}
	}
	return out
}

func IndexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	return slices.IndexFunc(s, f)
}

func Clone[S ~[]E, E any](s S) S {
	return slices.Clone(s)
}
