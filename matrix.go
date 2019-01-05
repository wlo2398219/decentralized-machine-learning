package main

import (
	"fmt"
	"log"
	"math"
)

type Matrix struct {
	row, col int
	mat      [][]float64
}

func (m1 *Matrix) mul(m2 *Matrix) *Matrix {

	if m1.col != m2.row {
		log.Fatal("INCONSISTENCY OF DIMENSION in mul m1:", m1.row, "x", m1.col, "m2:", m2.row, "x", m2.col)
	}

	var result = &Matrix{row: m1.row, col: m2.col, mat: make([][]float64, m1.row)}

	for i := 0; i < m1.row; i++ {
		result.mat[i] = make([]float64, m2.col)
	}

	for i := 0; i < m1.row; i++ {
		for j := 0; j < m2.col; j++ {
			result.mat[i][j] = innerProduct(m1.rowAt(i), m2.colAt(j))
		}
	}

	return result
}

func (m1 *Matrix) sub(m2 *Matrix) *Matrix {
	if m1.row != m2.row || m1.col != m2.col {
		log.Fatal("INCONSISTENCY OF DIMENSION in substraction, m1 = ", m1.row, "*", m1.col, ", m2 = ", m2.row, "*", m2.col)
	}

	var result = m1.getCopy()
	// var result = &Matrix{row: m1.row, col: m1.col, mat: m1.mat}

	for i := range result.mat {
		for j := range result.mat[0] {
			result.mat[i][j] -= m2.mat[i][j]
		}
	}

	return result
}

func (m1 *Matrix) add(m2 *Matrix) *Matrix {
	if m1.row != m2.row || m1.col != m2.col {
		log.Fatal("INCONSISTENCY OF DIMENSION in add in substraction")
	}
	var result = m1.getCopy()
	// var result = &Matrix{row: m1.row, col: m1.col, mat: m1.mat}

	for i := range result.mat {
		for j := range result.mat[0] {
			result.mat[i][j] += m2.mat[i][j]
		}
	}

	return result
}

func (m *Matrix) T() *Matrix {
	var result = Matrix{row: m.col, col: m.row, mat: make([][]float64, m.col)}

	for i := 0; i < m.col; i++ {
		result.mat[i] = make([]float64, m.row)
	}

	for i := 0; i < m.col; i++ {
		for j := 0; j < m.row; j++ {
			result.mat[i][j] = m.mat[j][i]
		}
	}

	return &result
}

func (m *Matrix) norm(p float64) float64 {
	ans := 0.0
	for i := 0; i < m.row; i++ {
		for j := 0; j < m.col; j++ {
			ans += math.Pow(m.mat[i][j], p)
		}
	}
	return math.Pow(ans, 1/p)
}

func (m *Matrix) getCopy() *Matrix {
	var result = &Matrix{row: m.row, col: m.col, mat: make([][]float64, m.row)}

	for i := range result.mat {
		result.mat[i] = make([]float64, m.col)
		copy(result.mat[i], m.mat[i])
	}

	return result
}

func (m *Matrix) sigmoid() *Matrix {

	var result = m.getCopy()
	// var result = &Matrix{row: m.row, col: m.col, mat: make([][]float64, m.row)}

	for i := range m.mat {
		// result.mat[i] = make([]float64, m.col)
		// copy(result.mat[i], )
		for j := range m.mat[0] {
			result.mat[i][j] = 1.0 / (1 + math.Exp(-result.mat[i][j]))
		}
	}

	return result
}

func (m *Matrix) log() *Matrix {

	var result = m.getCopy()
	// var result = &Matrix{row: m.row, col: m.col, mat: make([][]float64, m.row)}

	for i := range m.mat {
		// result.mat[i] = make([]float64, m.col)
		for j := range m.mat[0] {
			result.mat[i][j] = math.Log(result.mat[i][j])
		}
	}

	return result
}

func (m *Matrix) addConstant(c float64) *Matrix {

	var result = m.getCopy()
	// var result = &Matrix{row: m.row, col: m.col, mat: make([][]float64, m.row)}

	for i := range m.mat {
		// result.mat[i] = make([]float64, m.col)
		for j := range m.mat[0] {
			result.mat[i][j] += c
		}
	}

	return result
}

func (m *Matrix) mulConstant(c float64) *Matrix {

	var result = m.getCopy()
	// var result = &Matrix{row: m.row, col: m.col, mat: make([][]float64, m.row)}

	for i := range m.mat {
		// result.mat[i] = make([]float64, m.col)
		for j := range m.mat[0] {
			result.mat[i][j] *= c
		}
	}

	return result
}

func (m *Matrix) standardize() *Matrix {

	var (
		N      = float64(m.row)
		result = m.getCopy()
		mean   float64
		sigma  float64
	)

	// fmt.Println("N:", N)

	for j := 0; j < m.col; j++ {
		tmp := float64(0)
		for i := 0; i < m.row; i++ {
			tmp += result.mat[i][j]
		}

		mean = tmp / N

		tmp = 0
		for i := 0; i < m.row; i++ {
			tmp += math.Pow(result.mat[i][j]-mean, 2)
		}

		// fmt.Println(tmp)

		sigma = math.Sqrt(tmp / N)

		for i := 0; i < m.row; i++ {
			result.mat[i][j] = (result.mat[i][j] - mean) / sigma
		}
	}

	return result

}

func (m1 *Matrix) colAt(ind int) []float64 {
	if m1.col <= ind {
		log.Fatal("EXCEED DIMENSION, m1:", m1.col, "REQUESTED:", ind)
	}

	var result = make([]float64, m1.row)

	for i := 0; i < m1.row; i++ {
		result[i] = m1.mat[i][ind]
	}

	return result
}

func (m1 *Matrix) rowAt(ind int) []float64 {
	if m1.row <= ind {
		log.Fatal("EXCEED DIMENSION, m1:", m1.row, "REQUESTED:", ind)
	}

	var result = make([]float64, m1.col)

	for i := 0; i < m1.col; i++ {
		result[i] = m1.mat[ind][i]
	}

	return result
}

func (m *Matrix) print() {

	for i := 0; i < m.row; i++ {
		fmt.Print(m.mat[i][0])
		for j := 1; j < m.col; j++ {
			fmt.Print(" ", m.mat[i][j])
		}
		fmt.Println()
	}
	fmt.Println()
}

func innerProduct(v1 []float64, v2 []float64) float64 {
	var result float64
	var dim = len(v1)

	result = 0

	for i := 0; i < dim; i++ {
		result += v1[i] * v2[i]
	}

	return result
}

func sliceToMat(v []float64) *Matrix {
	N := len(v)
	result := &Matrix{row: N, col: 1, mat: make([][]float64, N)}

	for i, val := range v {
		result.mat[i] = make([]float64, 1)
		result.mat[i][0] = val
	}

	return result
}

func getZeroMat(r, c int) *Matrix {
	result := &Matrix{row: r, col: c, mat: make([][]float64, r)}

	for i := 0; i < r; i++ {
		result.mat[i] = make([]float64, c)
	}

	return result
}

func initMatrix(data [][]float64) *Matrix {
	r, c := len(data), len(data[0])
	result := &Matrix{row: r, col: c, mat: data}

	return result
}

func diagMatrix(data []float64) *Matrix {
	size := len(data)
	result := &Matrix{row: size, col: size, mat: make([][]float64, size)}

	for i := 0; i < size; i++ {
		result.mat[i] = make([]float64, size)
		result.mat[i][i] = data[i]
	}

	return result
}
