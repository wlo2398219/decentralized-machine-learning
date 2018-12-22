package main

import (
	"log"
	"strconv"
	"strings"
	// "fmt"
	"io/ioutil"
)

func load_data(filename string) (featureType, weightType) {
	var (
		x [][]float64
		y []float64
		w []float64
	)

	switch filename {
	case "uci_cbm_dataset.txt":
		x, y, w = uci_cbm()
	}

	return featureType{val: x, output: y}, weightType{val: w}
}

func uci_cbm() ([][]float64, []float64, []float64) {
	var (
		N, nf    = 11934, 15
		x        = make([][]float64, N)
		y        = make([]float64, N)
		w        = make([]float64, 15)
		data     = make([]float64, nf)
		filename = "./_Datasets/uci_cbm_dataset.txt"
	)

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal("Failed to open " + filename)
	}

	lines := strings.Split(string(b), "\n")

	for ind, line := range lines[:11934] {

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
