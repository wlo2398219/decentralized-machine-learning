package main 

import (
	// "log"
	"math"
	"math/rand"
	// "fmt"
)

type NNLayer interface {
	forward(*Matrix) *Matrix
	backward(*Matrix) *Matrix
	updatePara(float64)
}

type SoftmaxLayer struct {
	X *Matrix
	Y *Matrix
}

type CrossEntropyLayer struct {
	X *Matrix
	Y []int
}

type FCLayer struct {
	X *Matrix
	Y *Matrix
	W *Matrix
	B *Matrix
	DW *Matrix
	DB *Matrix
}

type MSELoss struct {
	X *Matrix
	Y *Matrix
	// Y []float64
}


func (l *MSELoss) forward(x *Matrix, y *Matrix) float64 {

	tmp := math.Pow(x.sub(y).norm(2), 2)

	l.X = x
	l.Y = y

	return tmp/float64(x.row)
}

func (l *MSELoss) backward() *Matrix {
	return l.X.sub(l.Y).mulConstant(1/float64(l.X.row))
}


func newLinearLayer(input, units int) FCLayer{
	
	w := &Matrix{row: units, col: input, mat: make([][]float64, units)}
	b := &Matrix{row: units, col: 1, mat: make([][]float64, units)}
	stdv := 1.0 / math.Sqrt(float64(input))
	dw := getZeroMat(units, input)
	db := getZeroMat(units, 1)

	for i := 0 ; i < w.row ; i++ {
		w.mat[i] = make([]float64, input)

		for j := 0 ; j < w.col ; j++ {
			w.mat[i][j] = (rand.Float64() - 0.5) * 2 * stdv
		}

		b.mat[i] = make([]float64, 1)
		b.mat[i][0] = (rand.Float64() - 0.5) * 2 * stdv
	}
	
	return FCLayer{W:w, B:b, DW:dw, DB:db}

}

// x: N x d
// w: unit x input
// b: unit x 1
func (l *FCLayer) forward(x *Matrix) *Matrix {
	result := x.mul(l.W.T())

	for i := 0 ; i < result.row ; i ++ {
		for j := 0 ; j < result.col ; j++ {
			result.mat[i][j] += l.B.mat[j][0]
		}
	}

	l.X = x
	l.Y = result

	return result
}

func (l *FCLayer) backward(dz *Matrix) *Matrix {

	zt := dz.T()
	tmp := &Matrix{row: zt.row, col: 1, mat: make([][]float64, zt.row)}

	for i := 0 ; i < zt.row ; i++ {
		tmp.mat[i] = make([]float64, 1)

		for j := 0 ; j < zt.col ; j ++ {
			tmp.mat[i][0] += zt.mat[i][j]
		}

	}	

	l.DW = l.DW.add(dz.T().mul(l.X))
	l.DB = l.DB.add(tmp)

	return dz.mul(l.W)
}

func (l *FCLayer) updatePara(gamma float64) {
	l.W = l.W.sub(l.DW.mulConstant(gamma))
	l.B = l.B.sub(l.DB.mulConstant(gamma))
	l.DW = getZeroMat(l.DW.row, l.DW.col)
	l.DB = getZeroMat(l.DB.row, 1)
}

// x: N x d
// output: N x d
func (l *SoftmaxLayer) forward(x *Matrix) *Matrix {

	result := x.getCopy()

	for i := 0 ; i < x.row ; i ++ {
		sum := float64(0)

		for j := 0 ; j < x.col ; j++ {
			tmp := math.Exp(x.mat[i][j])
			sum += tmp
			result.mat[i][j] = tmp
		}

		// test := float64(0)
		for j := 0 ; j < x.col ; j++ {
			result.mat[i][j] /= sum
			// test += result.mat[i][j]
		}

		// fmt.Println()
	}

	l.X = x
	l.Y = result

	return result

}

// dz : N x c
// x: N x c 
func (l *SoftmaxLayer) backward(dz *Matrix) *Matrix {
	
	result := getZeroMat(dz.row, dz.col)
	
	for i := 0 ; i < dz.row ; i++ {

		vecZ := getZeroMat(dz.col, 1)
		vecY := getZeroMat(dz.col, 1)

		diag := diagMatrix(l.Y.mat[i])

		for j := 0 ; j < dz.col ; j++ {
			vecZ.mat[j][0] = dz.mat[i][j]
			vecY.mat[j][0] = l.Y.mat[i][j]
		}

		// fmt.Println(vecY.row, vecY.col)
		// fmt.Println(vecZ.row, vecZ.col)
		// fmt.Println(diag.row, diag.col)

		tmp := vecY.mul(vecY.T().mulConstant(-1.0)).add(diag).mul(vecZ)
		// tmp.print()

		for j := 0 ; j < dz.col ; j++ {
			result.mat[i][j] = tmp.mat[j][0]
		}

	}	

	// fmt.Println(result.row, result.col)

	return result

}


// x: N x c
// output: real value
func (l *CrossEntropyLayer) forward(x *Matrix, y []int) float64 {
	var (
		loss float64
		tmp = x.getCopy().log()
	)

	for i := 0 ; i < tmp.row ; i++ {
		loss -= tmp.mat[i][y[i]]
	}

	l.Y = y
	l.X = x

	return loss/float64(x.row)
}

// output: N x c
func (l *CrossEntropyLayer) backward() *Matrix {
	N := float64(l.X.row)
	result := getZeroMat(l.X.row, l.X.col)

	for i := 0 ; i < result.row ; i++ {
		result.mat[i][l.Y[i]] = -1.0/(l.X.mat[i][l.Y[i]] * N) // -yi/pi
	}

	return result
}
