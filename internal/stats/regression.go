package stats

// LinearSlope returns the least-squares slope of the (x, y) pairs.
// Returns 0 if there are fewer than 2 points or if x has no variance.
func LinearSlope(xs, ys []float64) float64 {
	n := len(xs)
	if n < 2 || n != len(ys) {
		return 0
	}

	var sumX, sumY, sumXY, sumXX float64
	fn := float64(n)
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumXX += xs[i] * xs[i]
	}

	denom := fn*sumXX - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (fn*sumXY - sumX*sumY) / denom
}
