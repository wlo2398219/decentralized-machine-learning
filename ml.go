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
// func SGD (f func(FeatureType, WeightType) OutputType,
// 		  df func(FeatureType, WeightType) WeightType,
// 		  init_weight WeightType) {

// }

type WeightType struct {
	Val []float64
}

type FeatureType struct {
	Val    [][]float64
	Output []float64
}

type OutputType struct {
	Val float64
}

type WeightPacket struct {
	Org    string
	IterID int
	Weight *WeightType
}

type GradientPacket struct {
	Org      string
	Dst      string
	IterID   int
	Gradient WeightType
}

// handle the weight packet received from gossip port
func handleWeight(conn *net.UDPConn, packet *WeightPacket) {
	// sendWeight
	// sendGradient
	key := packet.Org + strconv.Itoa(packet.IterID)

	if _, exist := nameIDTable[key]; exist {
		return
	} else {
		nameIDTable[key] = true
		// broadcast
		broadcastWeight(conn, packet)
		// fmt.Println("=== MY WEIGHT ===")
		grad := grad_f(feature, weight, "mse", "", 0)
		sendGradient(conn, &GradientPacket{Org: *name, Dst: packet.Org, IterID: packet.IterID, Gradient: grad}, packet.Org)
	}

}

// handle the Gradient packet received from gossip port
func handleGradient(conn *net.UDPConn, packet *GradientPacket) {
	// if myself -> update
	// esle -> sendGradient

	if packet.Dst == *name {
		// receive the packet
		gradCh <- packet
	} else {
		sendGradient(conn, packet, packet.Dst)
	}

}

// broadcast the weight to other peers, used by the one announcing the training
func broadcastWeight(conn *net.UDPConn, packet *WeightPacket) {
	// not shared

	// packetBytes := make([]byte, MAX_PACKET_SIZE)
	packet1 := GossipPacket{WeightPacket: packet}
	// fmt.Println("=============================")
	// fmt.Println(packet1)
	// fmt.Println(packet1.WeightPacket)
	// fmt.Println("=============================")
	// tmp_weight := WeightPacket{Org: packet.Org, IterID: packet.IterID, weight: packet.weight}
	// packet1 := GossipPacket{WeightPacket: &tmp_weight}
	for peer := range peer_list.Iter() {
		_ = sendPacketToAddr(conn, packet1, peer)
	}

	// packetBytes, _ = protobuf.Encode(packet1)

	// fmt.Println("!", packet1)
	// fmt.Println(packet1.WeightPacket)
	// fmt.Println(packetBytes)
	// sendToPeers(conn, "", packetBytes)

}

// send the Gradient to the host, used by the peer receiving the weight
func sendGradient(conn *net.UDPConn, packet *GradientPacket, Dst string) {
	// not shared
	// private message
	gossipPacket := GossipPacket{GradientPacket: packet}

	if _, exist := nextHopTable[Dst]; exist {
		_ = sendPacketToAddr(conn, gossipPacket, nextHopTable[Dst].NextHop)
	} else {
		fmt.Println("==== DON'T KNOW THE DESTINATION OF THE GRADIENT ====")
	}

}

// func newTrainig() chan<- *GradientPacket {
func newTraining(conn *net.UDPConn) {
	// load dataset
	gradCh = make(chan *GradientPacket)

	fmt.Println("MY INIT WEIGHTS")
	fmt.Println(weight.Val)
	go func() {
		// for iteration
		// for select <- ch
		k, d := 5, len(weight.Val)
		gamma := 0.1

		for round := 0; round < 10; round++ {

			broadcastWeight(conn, &WeightPacket{Org: *name, IterID: round, Weight: &weight})
			fmt.Println("====== TRAINING EPOCH", round, "======")

			// used to save sum(grad)
			updates := make([]float64, d)

			for i := 0; i < k; {

				select {

				case ch := <-gradCh:

					if ch.IterID == round { // same round

						i++

						fmt.Println("===== GET GRADIENT FROM", ch.Org, "=====")
						fmt.Println(ch.Gradient.Val)
						for i := range updates {
							updates[i] += ch.Gradient.Val[i]
						}

					}
				}
			}

			// update the weight
			for i := range updates {
				weight.Val[i] = weight.Val[i] - gamma*updates[i]/float64(k)
			}

			fmt.Println("CURRENT WEIGHTS:", weight.Val)

			loss := f(feature, weight, "mse", "", 0)
			fmt.Println("LOSS:", loss)

		}	

	}()

	// return gradCh
}

// f: loss function
func f(x FeatureType, w WeightType, loss_type, regularization string, lambda float64) OutputType {

	// CALCULATE LOSS (DEFAULT IS 2-NORM AND W/O REGULARIZATION)

	if len(x.Val[0]) != len(w.Val) {
		log.Fatal("INCONSISTENCY OF DIMENSION IN f")
	}

	// transform x, w, y to matrices
	var (
		row, col = len(x.Val), len(x.Val[0])
		mat_x    = &Matrix{row: row, col: col, mat: x.Val}
		mat_w    = &Matrix{row: col, col: 1, mat: make([][]float64, col)}
		mat_y    = &Matrix{row: row, col: 1, mat: make([][]float64, row)}
		loss     OutputType
	)

	for i := 0; i < col; i++ {
		mat_w.mat[i] = make([]float64, 1)
		mat_w.mat[i][0] = w.Val[i]
	}

	for i := 0; i < row; i++ {
		mat_y.mat[i] = make([]float64, 1)
		mat_y.mat[i][0] = x.Output[i]
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
		loss.Val += innerProduct(w.Val, w.Val)
	case "lasso":
		for i := 0; i < len(w.Val); i++ {
			loss.Val += math.Abs(w.Val[i])
		}
	default:

	}

	return loss
}
func grad_f(x FeatureType, w WeightType, loss_type, regularization string, lambda float64) WeightType {

	if len(x.Val[0]) != len(w.Val) {
		log.Fatal("INCONSISTENCY OF DIMENSION IN f")
	}

	// transform x, w, y to matrices
	var (
		row, col = len(x.Val), len(x.Val[0])
		mat_x    = &Matrix{row: row, col: col, mat: x.Val}
		mat_w    = &Matrix{row: col, col: 1, mat: make([][]float64, col)}
		mat_y    = &Matrix{row: row, col: 1, mat: make([][]float64, row)}
		grad     WeightType
	)

	for i := 0; i < col; i++ {
		mat_w.mat[i] = make([]float64, 1)
		mat_w.mat[i][0] = w.Val[i]
	}

	for i := 0; i < row; i++ {
		mat_y.mat[i] = make([]float64, 1)
		mat_y.mat[i][0] = x.Output[i]
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
func grad_mse(x *Matrix, w *Matrix, y *Matrix, regularization string, lambda float64) WeightType {
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

	return WeightType{Val: result}
}

// f(x,w,y) = (xw-y)^T(xw-y), x:nxd, w:dx1
func mse_loss(x *Matrix, w *Matrix, y *Matrix, regularization string, lambda float64) OutputType {
	N := float64(x.row)
	tmp := x.mul(w).sub(y)
	norm := float64(0)

	switch regularization {
	case "ridge":
		norm = (w.T().mul(w)).mat[0][0]
	case "lasso":
		for _, Val := range w.mat {
			norm += math.Abs(Val[0])
		}
	}

	return OutputType{Val: tmp.T().mul(tmp).mat[0][0]/(2*N) + lambda*norm}
}

func logistic_loss(x *Matrix, w *Matrix, y *Matrix) OutputType {
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

	return OutputType{Val: -loss.mat[0][0] / N}
}
