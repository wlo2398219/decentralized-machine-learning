package main

import (
	"fmt"
	// "github.com/dedis/protobuf"
	"log"
	"math"
	"net"
	"strconv"
	"time"
	"math/rand"
)

var (
	gradCh          chan *GradientPacket
	nameIDTable     = map[string]bool{}                // used to record if I receive the weight from this peer in the given round
	dataset string
	feature FeatureType
	globalWeight WeightType
	// mnistFeature, mnistWeight = mnist_dataset()
	fcLayer = newLinearLayer(500, 10)
	smLayer = SoftmaxLayer{}  
	ceLayer = CrossEntropyLayer{}
	globalX *Matrix
	globalY []int

	SAMPLE = 1000
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
	Dataset   string
	IterID int
	Weight *WeightType
}

type GradientPacket struct {
	Org      string
	Dst      string
	Dataset     string
	IterID   int
	Gradient WeightType
}

// handle the weight packet received from gossip port
func handleWeight(conn *net.UDPConn, packet *WeightPacket) {
	// sendWeight
	// sendGradient

	var (
		grad WeightType
		// matX *Matrix
		// Y []int
	)

	// fmt.Println("==== HANDLE WEIGHT PACKET", packet, " ====")
	// fmt.Println("ORG:",packet.Org)
	// fmt.Println("DATANAME:",packet.Dataset)
	// fmt.Println("ID:",packet.IterID)
	// fmt.Println("WEIGHT:",packet.Weight)

	fmt.Println("HANDLE WEIGHT IN ROUND", packet.IterID)

	if packet.Org == *name {
		return 
	}

	key := packet.Org + strconv.Itoa(packet.IterID)

	if dataset == "" {
		if packet.Dataset != "mnist" {
			feature, _ = load_data(packet.Dataset)
			fmt.Println("KIDDING ME???", packet.Dataset)
		} else {
			globalX, globalY = mnist_dataset(SAMPLE)
			fmt.Println("==== load mnist ====")
			// fmt.Println()
		}

		dataset = packet.Dataset
	} else {
		fmt.Println("HANDLE WEIGHT, DATASET:", dataset)
	}

	if _, exist := nameIDTable[key]; exist {
		return
	} else {
		nameIDTable[key] = true
		// broadcast

		broadcastWeight(conn, packet)

		if packet.Dataset != "mnist" {
			grad = grad_f(feature, *packet.Weight, "mse", "", 0)
		} else {
			fmt.Println("==== grad_f_nn ====")
			grad = grad_f_nn(*packet.Weight, globalX, globalY)
		}

		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)		
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
	// fmt.Println("===== IN broadcastWeight =====")
	// fmt.Println("ORG:",packet.Org)
	// fmt.Println("DATANAME:",packet.Dataset)
	// fmt.Println("ID:",packet.IterID)
	// fmt.Println("WEIGHT:",packet.Weight)
	packet1 := GossipPacket{WeightPacket: packet}
	for peer := range peer_list.Iter() {
		fmt.Println("==== BROADCAST WEIGHT TO", peer, " ====")
		_ = sendPacketToAddr(conn, packet1, peer)
	}

}

// send the Gradient to the host, used by the peer receiving the weight
func sendGradient(conn *net.UDPConn, packet *GradientPacket, Dst string) {
	// not shared
	// private message
	gossipPacket := GossipPacket{GradientPacket: packet}

	if Dst == *name {
		return
	}

	if _, exist := nextHopTable[Dst]; exist {
		
		_ = sendPacketToAddr(conn, gossipPacket, nextHopTable[Dst].NextHop)
		fmt.Println("SEND GRADIENT TO", Dst, "FROM", packet.Org, "IN ROUND", packet.IterID)
	} else {
		fmt.Println("==== DON'T KNOW THE DESTINATION OF THE GRADIENT ====", Dst)
	}

}

// func newTrainig() chan<- *GradientPacket {

func newTraining(conn *net.UDPConn, dataName string, ch chan *GossipPacket) {
	// load dataset
	gradCh = make(chan *GradientPacket)

	dataset = dataName

	if dataName != "mnist" {
		feature, _ = load_data(dataName)
	} else {
		globalX, globalY = mnist_dataset(SAMPLE)
	}


	// fmt.Println("MY INIT WEIGHTS")
	// fmt.Println(weight.Val)

	// go byzantineSGD(conn, dataName)
	go distributedSGD(conn, dataName)
}

func distributedSGD(conn *net.UDPConn, dataName string) {

	// for iteration
	// for select <- ch
	// the dataset we use for testing now: feature contains X(Val) and Y(Output), weight is a 0-vector
	var (
		weight WeightType
		matX, testMatX *Matrix
		Y, testY []int
		gamma float64
		grad WeightType
	)

	if dataName != "mnist" {
		fmt.Println("==== LOAD", dataName, "====")
		feature, weight = load_data(dataName)
	} else {
		matX, Y = mnist_dataset(SAMPLE)
		testMatX, testY = mnist_dataset_test(SAMPLE)
		fcLayer = newLinearLayer(500, 10)
		weight = flattenWB(fcLayer.W, fcLayer.B)
		// fmt.Println(" =====  mnist ===== ")
		// matX.print()
	}

	k, d := 3, len(weight.Val) // k is #weights to be got, d is the dimension of weight
	// k := 1 // k is #weights to be got, d is the dimension of weight
	
	if dataName != "mnist" {
		gamma = 0.0000000001      // gamma is learning step size
	} else {
		gamma = 5 * 1e-3     // gamma is learning step size
	}

	fcLayer.W, fcLayer.B = deFlatten(weight.Val)
	nnOutput := fcLayer.forward(matX)
	smOutput := smLayer.forward(nnOutput)
	loss_train := ceLayer.forward(smOutput, Y)
	fmt.Println("INIT LOSS in mnist:", loss_train) //, ", TEST LOSS:", loss_test)

	fmt.Println("DATASET:   ", dataName)
	for round := 0; round < 30; round++ {

		broadcastWeight(conn, &WeightPacket{Org: *name, IterID: round, Dataset:dataName, Weight: &weight})
		
		fmt.Println("====== TRAINING EPOCH", round, "======")

		// used to save sum(grad)
		// updates := make([]float64, d)

		// if dataName != "mnist" {
		// 	grad = grad_f(feature, weight, "mse", "", 0)
		// } else {
		// 	fmt.Println("==== grad_f_nn ====")
		// 	grad = grad_f_nn(weight, globalX, globalY)
		// }

		// updates := grad.Val

		updates := make([]float64, d)


		if dataName != "mnist" {
			grad = grad_f(feature, weight, "mse", "", 0)
		} else {
			fmt.Println("==== grad_f_nn ====")
			grad = grad_f_nn(weight, globalX, globalY)
		}

		go func(r int) {
			time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)		
			packet := GradientPacket{Org:*name, Dst:*name, Dataset:dataName, IterID: r, Gradient: grad}
			gradCh <- &packet
		}(round)

		for i := 0; i < k; {
			fmt.Println("WAIT FOR GRAIDENT...")
			select {

			case ch := <-gradCh:

				if ch.IterID == round { // same round

					i++

					fmt.Println("===== GET GRADIENT FROM", ch.Org, "=====")
					// fmt.Println(ch.Gradient.Val)
					for j := range updates {
						updates[j] += ch.Gradient.Val[j]
					}

				}
			}
		}

		fmt.Println("ENOUGH GRADIENT")
		// update the weight
		for i := range updates {
			weight.Val[i] = weight.Val[i] - gamma*updates[i]/float64(k)
		}
		globalWeight = weight

		// fmt.Println("CURRENT WEIGHTS:", weight.Val)

		if dataName != "mnist" {
			loss_train := f(feature, weight, "mse", "", 0)
			fmt.Println("LOSS:", loss_train) //, ", TEST LOSS:", loss_test)

		} else {
			// fmt.Println("dimension:", matX.row, matX.col)
			fcLayer.W, fcLayer.B = deFlatten(weight.Val)
			// nnOutput := fcLayer.forward(matX)
			nnOutput := fcLayer.forward(testMatX)
			smOutput := smLayer.forward(nnOutput)
			// loss_train := ceLayer.forward(smOutput, Y)
			loss_train := ceLayer.forward(smOutput, testY)
			fmt.Println("LOSS in mnist:", loss_train) //, ", TEST LOSS:", loss_test)

			count := 0
			for i := 0 ; i < SAMPLE ; i++ {
				max_ := nnOutput.mat[i][0]
				pred := 0
				for j := 1 ; j < 10 ; j++ {
					if max_ < nnOutput.mat[i][j] {
						max_ = nnOutput.mat[i][j]
						pred = j
					}
				}
				if pred == testY[i] {
					count += 1
				}
			}

			fmt.Println("Acc:", float64(count)*100/float64(SAMPLE))
		}

		// loss_test  := f(feature_test,  weight, "mse", "", 0)
	}
}

func byzantineSGD(conn *net.UDPConn, dataName string) {

	// var weight WeightType
	var (
		weight WeightType
		matX, testMatX *Matrix
		Y, testY []int
		gamma float64
		grad WeightType
	)
	// feature, weight = load_data(dataName)

	if dataName != "mnist" {
		fmt.Println("==== LOAD", dataName, "====")
		feature, weight = load_data(dataName)
	} else {
		matX, Y = mnist_dataset(SAMPLE)
		testMatX, testY = mnist_dataset_test(SAMPLE)
		fcLayer = newLinearLayer(500, 10)
		weight = flattenWB(fcLayer.W, fcLayer.B)
	}


	// Parameter.
	byzF := 1

	// Internal states for filters.
	lastPeerWeights := make(map[string]WeightType)
	lastPeerGradients := make(map[string]WeightType)
	lastWeight := weight
	lastGradient := weight
	weightHistory := make([]WeightType, 0)
	recentPeers := make([]string, 0)

	// Dampening component as described in the paper sec 3.2.
	// dampening := func(delay float64) float64 { return math.Exp(-0.2 * delay) }
	dampening := func(delay float64) float64 { return 1.0/(delay+1.0) }

	// Lipschitz filter as described in the paper sec 3.1.
	lipschitzFilter := func(grad *GradientPacket) bool {
		ok := false
		cnt := 0
		gradientEvo := sliceToMat(grad.Gradient.Val).
			sub(sliceToMat(lastGradient.Val)).norm(2)
		modelEvo := sliceToMat(weight.Val).
			sub(sliceToMat(lastWeight.Val)).norm(2)
		if modelEvo < 1e-9 {
			ok = true
		}
		newLC := gradientEvo / (modelEvo + 1e-9)

		for peer, _ := range nextHopTable {
			lastPeerGradient, exist := lastPeerGradients[peer]
			if !exist {
				ok = true
				break
			}
			lastPeerWeight, exist := lastPeerWeights[peer]
			if !exist {
				ok = true
				break
			}

			peerGradientEvo := sliceToMat(grad.Gradient.Val).
				sub(sliceToMat(lastPeerGradient.Val)).norm(2)
			peerModelEvo := sliceToMat(weightHistory[grad.IterID].Val).
				sub(sliceToMat(lastPeerWeight.Val)).norm(2)

			if peerModelEvo < 1e-9 {
				ok = true
				break
			}
			peerLC := peerGradientEvo / (peerModelEvo + 1e-9)

			if peerLC < newLC {
				cnt++
			}
		}

		if cnt <= len(nextHopTable)-byzF {
			ok = true
		}

		return ok
	}

	// Frequency filter as described in the paper sec 3.1.
	frequencyFilter := func(grad *GradientPacket) bool {
		ok := true
		for _, peer := range recentPeers {
			if peer == grad.Org {
				ok = false
			}
		}
		return ok
	}


	m := 1
	d := len(weight.Val)  // k is #weights to be got, d is the dimension of weight
	// gamma := 0.0000000001 // gamma is learning step size

	if dataName != "mnist" {
		gamma = 0.0000000001      // gamma is learning step size
	} else {
		gamma = 5 * 1e-4     // gamma is learning step size
	}


	fcLayer.W, fcLayer.B = deFlatten(weight.Val)
	nnOutput := fcLayer.forward(matX)
	smOutput := smLayer.forward(nnOutput)
	loss_train := ceLayer.forward(smOutput, Y)
	fmt.Println("INIT LOSS in mnist:", loss_train) //, ", TEST LOSS:", loss_test)

	weightHistory = append(weightHistory, weight)
	for round := 0; round < 100; round++ {
		fmt.Println("====== TRAINING ITERATION", round, "======")
		broadcastWeight(conn, &WeightPacket{Org: *name, IterID: round, Weight: &weight, Dataset: dataName})

		// used to save sum(grad)
		updates := make([]float64, d)


		if dataName != "mnist" {
			grad = grad_f(feature, weight, "mse", "", 0)
		} else {
			fmt.Println("==== grad_f_nn ====")
			grad = grad_f_nn(weight, globalX, globalY)
		}

		go func(r int) {
			time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)		
			packet := GradientPacket{Org:*name, Dst:*name, Dataset:dataName, IterID: r, Gradient: grad}
			gradCh <- &packet
		}(round)
		// updates := grad.Val

		count := 0

		for count < m {
			grad := <-gradCh
			fmt.Println("===== GET GRADIENT FROM", grad.Org, "=====")
			// fmt.Println(grad.Gradient.Val, lipschitzFilter(grad), frequencyFilter(grad))
			if lipschitzFilter(grad) && frequencyFilter(grad) {
				delay := float64(round - grad.IterID)
				fmt.Println("round:", round, "IterID:", grad.IterID, "delay:", delay)

				for j := range updates {
					updates[j] += dampening(0) * grad.Gradient.Val[j]
					// updates[j] += grad.Gradient.Val[j]
				}

				count++

				// Update history list for the filters.
				lastPeerWeights[grad.Org] = weight
				lastPeerGradients[grad.Org] = grad.Gradient
				if len(recentPeers) == 2*byzF {
					recentPeers = recentPeers[1:]
				}
				recentPeers = append(recentPeers, grad.Org)
			}
		}

		// update the weight
		lastWeight = weight
		lastGradient.Val = updates
		for i := range updates {
			weight.Val[i] = weight.Val[i] - gamma*updates[i]
		}
		weightHistory = append(weightHistory, weight)
		globalWeight = weight

		// fmt.Println("CURRENT WEIGHTS:", weight.Val)

		if dataName != "mnist" {
			loss_train := f(feature, weight, "mse", "", 0)
			fmt.Println("LOSS:", loss_train) //, ", TEST LOSS:", loss_test)

		} else {
			// fmt.Println("dimension:", matX.row, matX.col)
			fcLayer.W, fcLayer.B = deFlatten(weight.Val)
			// nnOutput := fcLayer.forward(matX)
			nnOutput := fcLayer.forward(testMatX)
			smOutput := smLayer.forward(nnOutput)
			// loss_train := ceLayer.forward(smOutput, Y)
			loss_train := ceLayer.forward(smOutput, testY)
			fmt.Println("LOSS in mnist:", loss_train) //, ", TEST LOSS:", loss_test)

			count := 0
			for i := 0 ; i < SAMPLE ; i++ {
				max_ := nnOutput.mat[i][0]
				pred := 0
				for j := 1 ; j < 10 ; j++ {
					if max_ < nnOutput.mat[i][j] {
						max_ = nnOutput.mat[i][j]
						pred = j
					}
				}
				if pred == testY[i] {
					count += 1
				}
			}

			fmt.Println("Acc:", float64(count)*100/float64(SAMPLE))

		}

		// loss_train := f(feature, weight, "mse", "", 0)
		// // loss_test  := f(feature_test,  weight, "mse", "", 0)
		// fmt.Println("LOSS:", loss_train) //, ", TEST LOSS:", loss_test)
	}
}


func newTesting(conn *net.UDPConn, dataFilename string) {
	go func(){
		// Call python feature extractor.
		var dataFeature FeatureType
		weight := globalWeight
		dataFeature = extractFeature(dataFilename)

		// Pass the feature to model and get the output.
		if dataset != "mnist" {
			// TODO: Implement this if needed.
		} else {
			fcLayer.W, fcLayer.B = deFlatten(weight.Val)
			testMatX := sliceToMat(dataFeature.Val[0]).T()
			nnOutput := fcLayer.forward(testMatX)

			max_ := nnOutput.mat[0][0]
			pred := 0
			for j := 1 ; j < 10 ; j++ {
				if max_ < nnOutput.mat[0][j] {
					max_ = nnOutput.mat[0][j]
					pred = j
				}
			}

			fmt.Println("PRED: data =", dataFilename, "pred =", pred,
				"label =", dataFeature.Output[0])
		}

		// TODO (or not?): Display (print?) output.
	}()
}

// f: loss function
// lambda: for regularziation -> lambda * regularization
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

func grad_f_nn(w WeightType, matX *Matrix, Y []int) WeightType {
	var grad = WeightType{Val: make([]float64, len(w.Val))}
	ind := 0
	wx := &Matrix{row: 10, col: 500, mat: make([][]float64, 10)}
	bx := &Matrix{row: 10, col: 1, mat: make([][]float64, 10)}

	for i := 0 ; i < wx.row ; i++ {
		wx.mat[i] = make([]float64, wx.col)

		for j := 0 ; j < wx.col ; j++ {
			wx.mat[i][j] = w.Val[ind]
			ind++
		}
	}

	for i := 0 ; i < bx.row ; i++ {
		bx.mat[i] = make([]float64, bx.col)
		bx.mat[i][0] = w.Val[ind]
		ind++
	}

	fcLayer.W = wx
	fcLayer.B = bx
	fcLayer.DW = getZeroMat(wx.row, wx.col)
	fcLayer.DB = getZeroMat(bx.row, 1)

	// fmt.Println("===== matX =====")
	// matX.print()

	nnOutput := fcLayer.forward(matX)
	smOutput := smLayer.forward(nnOutput)
	_ = ceLayer.forward(smOutput, Y)

	back := ceLayer.backward()
	back = smLayer.backward(back)
	fcLayer.backward(back)

	ind = 0

	for i := 0 ; i < wx.row ; i++ {
		for j := 0 ; j < wx.col ; j++ {
			grad.Val[ind] = fcLayer.DW.mat[i][j]
			ind++
		}
	}

	for i := 0 ; i < bx.row ; i++ {
		grad.Val[ind] = fcLayer.DB.mat[i][0]
		ind++
	}

	return grad
}


func grad_f(x FeatureType, w WeightType, loss_type, regularization string, lambda float64) WeightType {

	// fmt.Println(x)
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


func flattenWB (w *Matrix, b *Matrix) WeightType {
	ind := 0
	weights := make([]float64, 5010)

	for i := 0 ; i < w.row ; i++ {
		for j := 0 ; j < w.col ; j++ {
			weights[ind] = w.mat[i][j]
			ind++
		}
	}

	for i := 0 ; i < b.row ; i++ {
		weights[ind] = b.mat[i][0]
		ind++
	}

	return WeightType{Val: weights}
}

func deFlatten(weights []float64) (*Matrix, *Matrix) {
	wx := getZeroMat(10, 500)
	bx := getZeroMat(10, 1)
	ind := 0

	for i := 0 ; i < wx.row ; i++ {
		for j := 0 ; j < wx.col ; j++ {
			wx.mat[i][j] = weights[ind]
			ind++
		}
	}

	for i := 0 ; i < bx.row ; i++ {
		bx.mat[i][0] = weights[ind]
		ind++
	}

	return wx, bx
}
