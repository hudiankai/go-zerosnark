package r1csqapFloat

import "math/big"

func Transpose(matrix [][]*big.Float) [][]*big.Float {
	var r [][]*big.Float
	for i := 0; i < len(matrix[0]); i++ {
		var row []*big.Float
		for j := 0; j < len(matrix); j++ {
			row = append(row, matrix[j][i])
		}
		r = append(r, row)
	}
	return r
}

func ArrayOfBigZeros(num int) []*big.Float {
	bigZero := big.NewFloat(float64(0))
	var r []*big.Float
	for i := 0; i < num; i++ {
		r = append(r, bigZero)
	}
	return r
}

func PolMul(a, b []*big.Float) []*big.Float {
	r := ArrayOfBigZeros(len(a) + len(b) - 1)
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			r[i+j] = new(big.Float).Add(
				r[i+j],
				new(big.Float).Mul(a[i], b[j]))
		}
	}
	return r
}

func PolDiv(a, b []*big.Float) ([]*big.Float, []*big.Float) {
	// https://en.wikipedia.org/wiki/Division_algorithm
	r := ArrayOfBigZeros(len(a) - len(b) + 1)
	rem := a
	for len(rem) >= len(b) {
		l := new(big.Float).Quo(rem[len(rem)-1], b[len(b)-1])
		pos := len(rem) - len(b)
		r[pos] = l
		aux := ArrayOfBigZeros(pos)
		aux1 := append(aux, l)
		aux2 := PolSub(rem, PolMul(b, aux1))
		rem = aux2[:len(aux2)-1]
	}
	return r, rem
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func PolAdd(a, b []*big.Float) []*big.Float {
	r := ArrayOfBigZeros(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = new(big.Float).Add(r[i], a[i])
	}
	for i := 0; i < len(b); i++ {
		r[i] = new(big.Float).Add(r[i], b[i])
	}
	return r
}

func PolSub(a, b []*big.Float) []*big.Float {
	r := ArrayOfBigZeros(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = new(big.Float).Add(r[i], a[i])
	}
	for i := 0; i < len(b); i++ {
		bneg := new(big.Float).Mul(b[i], big.NewFloat(float64(-1)))
		r[i] = new(big.Float).Add(r[i], bneg)
	}
	return r

}

func FloatPow(a *big.Float, e int) *big.Float {
	if e == 0 {
		return big.NewFloat(float64(1))
	}
	result := new(big.Float).Copy(a)
	for i := 0; i < e-1; i++ {
		result = new(big.Float).Mul(result, a)
	}
	return result
}

func PolEval(v []*big.Float, x *big.Float) *big.Float {
	r := big.NewFloat(float64(0))
	for i := 0; i < len(v); i++ {
		xi := FloatPow(x, i)
		elem := new(big.Float).Mul(v[i], xi)
		r = new(big.Float).Add(r, elem)
	}
	return r
}

func NewPolZeroAt(pointPos, totalPoints int, height *big.Float) []*big.Float {
	fac := 1
	for i := 1; i < totalPoints+1; i++ {
		if i != pointPos {
			fac = fac * (pointPos - i)
		}
	}
	facBig := big.NewFloat(float64(fac))
	hf := new(big.Float).Quo(height, facBig)
	r := []*big.Float{hf}
	for i := 1; i < totalPoints+1; i++ {
		if i != pointPos {
			ineg := big.NewFloat(float64(-i))
			b1 := big.NewFloat(float64(1))
			r = PolMul(r, []*big.Float{ineg, b1})
		}
	}
	return r
}

func LagrangeInterpolation(v []*big.Float) []*big.Float {
	// https://en.wikipedia.org/wiki/Lagrange_polynomial
	var r []*big.Float
	for i := 0; i < len(v); i++ {
		r = PolAdd(r, NewPolZeroAt(i+1, len(v), v[i]))
	}
	//
	return r
}

func R1CSToQAP(a, b, c [][]*big.Float) ([][]*big.Float, [][]*big.Float, [][]*big.Float, []*big.Float) {
	aT := Transpose(a)
	bT := Transpose(b)
	cT := Transpose(c)
	var alphas [][]*big.Float
	for i := 0; i < len(aT); i++ {
		alphas = append(alphas, LagrangeInterpolation(aT[i]))
	}
	var betas [][]*big.Float
	for i := 0; i < len(bT); i++ {
		betas = append(betas, LagrangeInterpolation(bT[i]))
	}
	var gammas [][]*big.Float
	for i := 0; i < len(cT); i++ {
		gammas = append(gammas, LagrangeInterpolation(cT[i]))
	}
	z := []*big.Float{big.NewFloat(float64(1))}
	for i := 1; i < len(aT[0])+1; i++ {
		ineg := big.NewFloat(float64(-i))
		b1 := big.NewFloat(float64(1))
		z = PolMul(z, []*big.Float{ineg, b1})
	}
	return alphas, betas, gammas, z
}

func CombinePolynomials(r []*big.Float, ap, bp, cp [][]*big.Float) ([]*big.Float, []*big.Float, []*big.Float, []*big.Float) {
	var alpha []*big.Float
	for i := 0; i < len(r); i++ {
		m := PolMul([]*big.Float{r[i]}, ap[i])
		alpha = PolAdd(alpha, m)
	}
	var beta []*big.Float
	for i := 0; i < len(r); i++ {
		m := PolMul([]*big.Float{r[i]}, bp[i])
		beta = PolAdd(beta, m)
	}
	var gamma []*big.Float
	for i := 0; i < len(r); i++ {
		m := PolMul([]*big.Float{r[i]}, cp[i])
		gamma = PolAdd(gamma, m)
	}

	px := PolSub(PolMul(alpha, beta), gamma)
	return alpha, beta, gamma, px
}

func DivisorPolinomial(px, z []*big.Float) []*big.Float {
	quo, _ := PolDiv(px, z)
	return quo
}
