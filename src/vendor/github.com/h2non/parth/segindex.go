package parth

func segStartIndexFromStart(path string, seg int) (int, bool) {
	if seg < 0 {
		return 0, false
	}

	for n, ct := 0, 0; n < len(path); n++ {
		if n > 0 && path[n] == '/' {
			ct++
		}

		if ct == seg {
			return n, true
		}
	}

	return 0, false
}

func segStartIndexFromEnd(path string, seg int) (int, bool) {
	if seg > -1 {
		return 0, false
	}

	for n, ct := len(path)-1, 0; n >= 0; n-- {
		if path[n] == '/' || n == 0 {
			ct--
		}

		if ct == seg {
			return n, true
		}
	}

	return 0, false
}

func segEndIndexFromStart(path string, seg int) (int, bool) {
	if seg < 1 {
		return 0, false
	}

	for n, ct := 0, 0; n < len(path); n++ {
		if path[n] == '/' && n > 0 {
			ct++
		}

		if ct == seg {
			return n, true
		}

		if n+1 == len(path) && ct+1 == seg {
			return n + 1, true
		}
	}

	return 0, false
}

func segEndIndexFromEnd(path string, seg int) (int, bool) {
	if seg > 0 {
		return 0, false
	}

	if seg == 0 {
		return len(path), true
	}

	if len(path) == 1 && path[0] == '/' {
		return 0, true
	}

	for n, ct := len(path)-1, 0; n >= 0; n-- {
		if n == 0 || path[n] == '/' {
			ct--
		}

		if ct == seg {
			return n, true
		}

	}

	return 0, false
}

func segIndexByKey(path, key string) (int, bool) { //nolint
	if path == "" || key == "" {
		return 0, false
	}

	for n := 0; n < len(path); n++ {
		si, ok := segStartIndexFromStart(path, n)
		if !ok {
			return 0, false
		}

		if len(path[si:]) == len(key)+1 {
			if path[si+1:] == key {
				return si, true
			}

			return 0, false
		}

		tmpEI, ok := segStartIndexFromStart(path[si:], 1)
		if !ok {
			return 0, false
		}

		if path[si+1:tmpEI+si] == key || n == 0 && path[0] != '/' && path[si:tmpEI+si] == key {
			return si, true
		}
	}

	return 0, false
}
