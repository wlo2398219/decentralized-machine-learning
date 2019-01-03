package main 

import (
	// "log"
	"math"
	"math/rand"
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
	Y *Matrix
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
func (l *SoftmaxLayer) forward(x *Matrix) *Matrix {

	result := x.getCopy()

	for i := 0 ; i < x.row ; i ++ {
		sum := float64(0)

		for j := 0 ; j < x.col ; j++ {
			tmp := math.Exp(x.mat[i][j])
			sum += tmp
			result.mat[i][j] = tmp
		}

		for j := 0 ; j < x.col ; j++ {
			result.mat[i][j] /= sum
		}

	}

	l.X = x
	l.Y = result

	return result

}

func (l *SoftmaxLayer) backward(z *Matrix) *Matrix {

	dim := l.X.row // #categories
	diag := &Matrix{row: dim, col: dim, mat: make([][]float64, dim)}
	result := l.Y.mul(l.Y.T())


	for i := 0 ; i < dim ; i++ {
		diag.mat[i] = make([]float64, dim)
		diag.mat[i][i] = l.Y.mat[i][0]
	}

	result = result.add(diag)

	return result.mul(z)
}


func (l *CrossEntropyLayer) forward(x *Matrix, y *Matrix) *Matrix {
	var (
		loss float64
		tmp = x.getCopy().log()
		result = &Matrix{row: 1, col: 1, mat: make([][]float64, 1)}
	)

	result.mat[0] = make([]float64, 1)

	for i := 0 ; i < tmp.row ; i++ {
		loss += -1.0 * tmp.mat[i][0] * y.mat[i][0]
	}

	result.mat[0][0] = loss
	l.Y = result
	l.X = x

	return result

}

func (l *CrossEntropyLayer) backward() *Matrix {

	// return result.mul(z)
	return &Matrix{}
}




