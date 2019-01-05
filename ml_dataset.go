package main

import (
	"log"
	"strconv"
	"strings"
	// "fmt"
	"io/ioutil"
)

func load_data(filename string) (FeatureType, WeightType) {
	var (
		x [][]float64
		y []float64
		w []float64
	)

	switch filename {
	case "uci_cbm_dataset.txt":
		x, y, w = uci_cbm()
	case "test":
		x, y, w = test_dataset()
	}

	return FeatureType{Val: x, Output: y}, WeightType{Val: w}
}

func uci_cbm() ([][]float64, []float64, []float64) {
	var (
		N, nf    = 11934, 15
		x        = make([][]float64, N)
		y        = make([]float64, N)
		w        = make([]float64, nf)
		data     = make([]float64, nf)
		filename = "./_Datasets/uci_cbm_dataset.txt"
	)

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal("Failed to open " + filename)
	}

	lines := strings.Split(string(b), "\n")

	for ind, line := range lines[:N] {

		tmp := strings.Split(line[3:], "   ")

		for i := 0; i < 8; i++ {
			data[i], _ = strconv.ParseFloat(tmp[i], 64)
		}

		for i := 9; i < 16; i++ {
			data[i-1], _ = strconv.ParseFloat(tmp[i], 64)
		}

		x[ind] = make([]float64, nf)
		copy(x[ind], data)
		y[ind], _ = strconv.ParseFloat(tmp[16], 64)

	}

	// fmt.Println(x[0], y[0])
	// fmt.Println(x[10], y[10])

	return x, y, w

}

func test_dataset() ([][]float64, []float64, []float64) {
	var (
		N, nf    = 1000, 16
		x        = make([][]float64, N)
		y        = make([]float64, N)
		w        = make([]float64, nf)
		data     = make([]float64, nf+1)
		filename = "./_Datasets/fake.txt"
	)

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal("Failed to open " + filename)
	}

	lines := strings.Split(string(b), "\n")

	for ind, line := range lines[:N] {

		tmp := strings.Split(line, ",")

		for i := 0; i < nf; i++ {
			data[i], _ = strconv.ParseFloat(tmp[i], 64)
		}

		x[ind] = make([]float64, nf)
		copy(x[ind], data)
		y[ind], _ = strconv.ParseFloat(tmp[nf], 64)

	}

	// fmt.Println(x[0], y[0])
	// fmt.Println(x[10], y[10])

	return x, y, w

}

func mnist_dataset(N int) (*Matrix, []int) {
	var (
		nf    = 500
		x        = make([][]float64, N)
		y        = make([]int, N)
		data     = make([]float64, nf)
		filename = "./_Datasets/hidden_layer_train.csv"
	)

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal("Failed to open " + filename)
	}

	lines := strings.Split(string(b), "\n")

	for ind, line := range lines[:N] {

		tmp := strings.Split(line, ",")
		tmp1, _ := strconv.ParseFloat(tmp[0], 64)
		y[ind] = int(tmp1)

		for i := 1; i < nf+1; i++ {
			data[i-1], _ = strconv.ParseFloat(tmp[i], 64)
		}

		x[ind] = make([]float64, nf)
		copy(x[ind], data)

	}
	
	matX := &Matrix{row: N, col: 500, mat: x}

	return matX, y

}
