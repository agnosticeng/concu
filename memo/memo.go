package memo

import "sync"

func Keyed[K comparable, V any](f func(key K) V) func(key K) V {
	var (
		m = make(map[K]V)
		l sync.Mutex
	)

	return func(key K) V {
		l.Lock()
		defer l.Unlock()

		v, ok := m[key]

		if ok {
			return v
		}

		v = f(key)
		m[key] = v
		return v
	}
}

func KeyedErr[K comparable, V any](f func(key K) (V, error)) func(key K) (V, error) {
	var (
		m = make(map[K]V)
		l sync.Mutex
	)

	return func(key K) (V, error) {
		l.Lock()
		defer l.Unlock()

		v, ok := m[key]

		if ok {
			return v, nil
		}

		v, err := f(key)

		if err != nil {
			return v, err
		}

		m[key] = v
		return v, nil
	}
}
