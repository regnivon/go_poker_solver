package main

import (
	"flag"
	"github.com/chehsunliu/poker"
	"log"
	"os"
	"quentin/solv"
	"runtime/pprof"
	//"fmt"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	} 


	//oopHands := "KQs, QJs, JTs, T9s, 98s, 77+"
	//ipHands := "KQs, QJs, JTs, T9s, 98s, 77+"
	board := []poker.Card{poker.NewCard("Ac"), poker.NewCard("7s"), poker.NewCard("5s"),
		poker.NewCard("3d"), poker.NewCard("2h")}

	oopHands := "AA, KK, QQ, JJ, TT, 99, 88, 77, 66, 55, 44, 33, 22, AK, AQ, AJ, AT, A9, A8, A7, A6, A5, A4, A3, A2, KQ, KJ, KT, K9, K8, K7, K6, K5, K4, K3, K2, QJ, QT, Q9, Q8, Q7, Q6, Q5, Q4, Q3, Q2, JT, J9, J8, J7, J6, J5, J4, J3, J2, T9, T8, T7, T6, T5, T4, T3, T2, 98, 97, 96, 95, 94, 93, 92, 87, 86, 85, 84, 83, 82, 76, 75, 74, 73, 72, 65, 64, 63, 62, 54, 53, 52, 43, 42, 32"
	ipHands := "AA, KK, QQ, JJ, TT, 99, 88, 77, 66, 55, 44, 33, 22, AK, AQ, AJ, AT, A9, A8, A7, A6, A5, A4, A3, A2, KQ, KJ, KT, K9, K8, K7, K6, K5, K4, K3, K2, QJ, QT, Q9, Q8, Q7, Q6, Q5, Q4, Q3, Q2, JT, J9, J8, J7, J6, J5, J4, J3, J2, T9, T8, T7, T6, T5, T4, T3, T2, 98, 97, 96, 95, 94, 93, 92, 87, 86, 85, 84, 83, 82, 76, 75, 74, 73, 72, 65, 64, 63, 62, 54, 53, 52, 43, 42, 32"


	//oopHands := "JJ"
	//ipHands := "QQ, T9s"


	oop := solv.HandsStringToHandRange(oopHands)
	ip := solv.HandsStringToHandRange(ipHands)

	oop = solv.RemoveConflicts(oop, board)
	ip = solv.RemoveConflicts(ip, board)

	bets := [][][]float64{
		{
			{
				1.0,
			},
		},
		{
			{
				1.0,
			},
		},
	}
	tree := solv.ConstructTree(40, 100, 1, bets, ip, oop)
	solv.OutputTree(tree, 0)
	traversal := solv.NewTraversal(oop, ip)
	solv.Train(traversal, 1000, tree)


	//average := tree.Strategy(0)//tree.GetAverageStrategy()
	//average2 := tree.GetNext(0).(*solv.GameNode).GetAverageStrategy()

	/*
	for hand := range oop {
		fmt.Println(oop[hand].Hand, tree.Strategy(hand))
	}
	for hand := range ip {
		
		fmt.Println(ip[hand].Hand, tree.GetNext(0).(*solv.GameNode).Strategy(hand))
	} */
	
	

	//fmt.Printf("%v\n", average)
	//fmt.Printf("%v\n", average2)
	//fmt.Printf("%v %v\n", len(oop), len(ip)) 
	//fmt.Println(betsum) 

}