package dstrfn

func ceilDiv(p, q int) int {
	switch {
	case p < 0 && q > 0:
		return -ceilDiv(-p, q)
	case p > 0 && q < 0:
		return -ceilDiv(p, -q)
	case p < 0 && q < 0:
		return ceilDiv(-p, -q)
	default:
		return (p + q - 1) / q
	}
}

func max(a, b int) int {
	switch {
	case b > a:
		return b
	default:
		return a
	}
}

func min(a, b int) int {
	switch {
	case b < a:
		return b
	default:
		return a
	}
}
