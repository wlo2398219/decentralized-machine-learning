package main

import (
	"fmt"
	// "github.com/dedis/protobuf"
	"log"
	"math"
	"net"
	"strconv"
)

var (
	gradCh          chan *GradientPacket
	nameIDTable     = map[string]bool{}
	feature, weight = load_data("uci_cbm_dataset.txt")
)

// Input
// func SGD (f func(featureType, weightType) outputType,
// 		  df func(featureType, weightType) weightType,
// 		  init_weight weightType) {

// }

type weightType struct {
	val []float64
}

type featureType struct {
	val    [][]float64
	output []float64
}

type outputType struct {
	val float64
}

type WeightPacket struct {
	org    string
	iterID int
	weight weightType
}

type GradientPacket struct {
	org      string
	dst      string
	iterID   int
	gradient weightType
}

// handle the weight packet received from gossip port
func handleWeight(conn *net.UDPConn, packet *WeightPacket) {
	// sendWeight
	// sendgradient
	key := packet.org + strconv.Itoa(packet.iterID)

	if _, exist := nameIDTable[key]; exist {
		return
	} else {
		nameIDTable[key] = true
		// broadcast
		broadcastWeight(conn, packet)
		grad := grad_f(feature, weight, "mse", "", 1.5)
		sendGradient(conn, &GradientPacket{org: *name, dst: packet.org, iterID: packet.iterID, gradient: grad}, packet.org)
	}

}

// handle the gradient packet received from gossip port
func handleGradient(conn *net.UDPConn, packet *GradientPacket) {
	// if myself -> update
	// esle -> sendGradient

	if packet.dst == *name {
		// receive the packet
		gradCh <- packet
	} else {
		sendGradient(conn, packet, packet.dst)
	}

}

// broadcast the weight to other peers, used by the one announcing the training
func broadcastWeight(conn *net.UDPConn, packet *WeightPacket) {
	// not shared

	// packetBytes := make([]byte, MAX_PACKET_SIZE)
	packet1 := GossipPacket{WeightPacket: packet}

	for peer := range peer_list.Iter() {
		_ = sendPacketToAddr(conn, packet1, peer)
	}

	// packetBytes, _ = protobuf.Encode(packet1)

	// fmt.Println("!", packet1)
	// fmt.Println(packet1.WeightPacket)
	// fmt.Println(packetBytes)
	// sendToPeers(conn, "", packetBytes)

}

// send the gradient to the host, used by the peer receiving the weight
func sendGradient(conn *net.UDPConn, packet *GradientPacket, dst string) {
	// not shared
	// private message
	gossipPacket := GossipPacket{GradientPacket: packet}

	if _, exist := nextHopTable[dst]; exist {
		_ = sendPacketToAddr(conn, gossipPacket, nextHopTable[dst].NextHop)
	} else {
		fmt.Println("==== DON'T KNOW THE DESTINATION OF THE GRADIENT ====")
	}

}

// func newTrainig() chan<- *GradientPacket {
func newTraining(conn *net.UDPConn) {
	// load dataset
	gradCh = make(chan *GradientPacket)

	go func() {
		// for iteration
		// for select <- ch
		k, d := 5, len(weight.val)
		gamma := 0.1

		for round := 0; round < 10; round++ {

			broadcastWeight(conn, &WeightPacket{org: *name, iterID: round, weight: weight})
			fmt.Println("====== TRAINING EPOCH", round, "======")

			// used to save sum(grad)
			updates := make([]float64, d)

			for i := 0; i < k; {

				select {

				case ch := <-gradCh:

					if ch.iterID == round { // same round

						i++

						fmt.Println("===== GET GRADIENT FROM", ch.org, "=====")

						for i := range updates {
							updates[i] += ch.gradient.val[i]
						}

					}
				}
			}

			// update the weight
			for i := range updates {
				weight.val[i] = weight.val[i] - gamma*updates[i]/float64(k)
			}

			loss := f(feature, weight, "mse", "", 1.5)
			fmt.Println("LOSS:", loss)

		}

	}()

	// return gradCh
}

// f: loss function
func f(x featureType, w weightType, loss_type, regularization string, lambda float64) outputType {

	// CALCULATE LOSS (DEFAULT IS 2-NORM AND W/O REGULARIZATION)

	if len(x.val[0]) != len(w.val) {
		log.Fatal("INCONSISTENCY OF DIMENSION IN f")
	}

	// transform x, w, y to matrices
	var (
		row, col = len(x.val), len(x.val[0])
		mat_x    = &Matrix{row: row, col: col, mat: x.val}
		mat_w    = &Matrix{row: col, col: 1, mat: make([][]float64, col)}
		mat_y    = &Matrix{row: row, col: 1, mat: make([][]float64, row)}
		loss     outputType
	)

	for i := 0; i < col; i++ {
		mat_w.mat[i] = make([]float64, 1)
		mat_w.mat[i][0] = w.val[i]
	}

	for i := 0; i < row; i++ {
		mat_y.mat[i] = make([]float64, 1)
		mat_y.mat[i][0] = x.output[i]
	}

	// compute loss
	switch loss_type {
	case "mse":
		loss = mse_loss(mat_x, mat_w, mat_y, regularization, lambda)
	case "logistic":
		loss = logistic_loss(mat_x, mat_w, mat_y)
	default:
		loss = mse_loss(mat_x, mat_w, mat_y, regularization, lambda)

	}

	switch regularization {
	case "ridge":
		loss.val += innerProduct(w.val, w.val)
	case "lasso":
		for i := 0; i < len(w.val); i++ {
			loss.val += math.Abs(w.val[i])
		}
	default:

	}

	return loss
}
func grad_f(x featureType, w weightType, loss_type, regularization string, lambda float64) weightType {

	if len(x.val[0]) != len(w.val) {
		log.Fatal("INCONSISTENCY OF DIMENSION IN f")
	}

	// transform x, w, y to matrices
	var (
		row, col = len(x.val), len(x.val[0])
		mat_x    = &Matrix{row: row, col: col, mat: x.val}
		mat_w    = &Matrix{row: col, col: 1, mat: make([][]float64, col)}
		mat_y    = &Matrix{row: row, col: 1, mat: make([][]float64, row)}
		grad     weightType
	)

	for i := 0; i < col; i++ {
		mat_w.mat[i] = make([]float64, 1)
		mat_w.mat[i][0] = w.val[i]
	}

	for i := 0; i < row; i++ {
		mat_y.mat[i] = make([]float64, 1)
		mat_y.mat[i][0] = x.output[i]
	}

	switch loss_type {
	case "mse":
		grad = grad_mse(mat_x, mat_w, mat_y, regularization, lambda)
	default:
		grad = grad_mse(mat_x, mat_w, mat_y, regularization, lambda)

	}

	return grad
}

// df = x^T(xw-y) + lambda * d(regularization)
func grad_mse(x *Matrix, w *Matrix, y *Matrix, regularization string, lambda float64) weightType {
	N := float64(x.row)
	grad := x.T().mul(x.mul(w).sub(y)).mulConstant(1.0 / N)
	result := make([]float64, w.row)

	switch regularization {
	case "ridge":
		grad = grad.add(w.mulConstant(lambda))
	case "lasso":
		tmp := w.getCopy()
		for i := range tmp.mat {
			if tmp.mat[i][0] >= 0 {
				tmp.mat[i][0] = 1.0
			} else {
				tmp.mat[i][0] = -1.0
			}
		}
		grad = grad.add(tmp.mulConstant(lambda))
	default:

	}

	for i := range grad.mat {
		result[i] = grad.mat[i][0]
	}

	return weightType{val: result}
}

// f(x,w,y) = (xw-y)^T(xw-y), x:nxd, w:dx1
func mse_loss(x *Matrix, w *Matrix, y *Matrix, regularization string, lambda float64) outputType {
	N := float64(x.row)
	tmp := x.mul(w).sub(y)
	norm := float64(0)

	switch regularization {
	case "ridge":
		norm = (w.T().mul(w)).mat[0][0]
	case "lasso":
		for _, val := range w.mat {
			norm += math.Abs(val[0])
		}
	}

	return outputType{val: tmp.T().mul(tmp).mat[0][0]/(2*N) + lambda*norm}
}

func logistic_loss(x *Matrix, w *Matrix, y *Matrix) outputType {
	N := float64(x.row)
	y_hat := x.mul(w).sigmoid()
	one_minus_y := &Matrix{row: y.row, col: 1, mat: make([][]float64, y.row)}
	one_minus_y_hat := &Matrix{row: y.row, col: 1, mat: make([][]float64, y.row)}

	for i := range one_minus_y_hat.mat {
		one_minus_y_hat.mat[i] = make([]float64, 1)
		one_minus_y.mat[i] = make([]float64, 1)

		copy(one_minus_y.mat[i], y.mat[i])
		copy(one_minus_y_hat.mat[i], y_hat.mat[i])
	}

	for i := range one_minus_y_hat.mat {
		one_minus_y_hat.mat[i][0] = 1 - one_minus_y_hat.mat[i][0]
	}

	loss := y.T().mul(y_hat.addConstant(1e-5).log())
	loss = loss.add(one_minus_y.T().mul(one_minus_y_hat.addConstant(1e-5).log()))

	fmt.Println("LOSS DIMENSION:", loss.row, loss.col)

	return outputType{val: -loss.mat[0][0] / N}
}
