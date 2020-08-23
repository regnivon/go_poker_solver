package main

import (
	"flag"
	"github.com/chehsunliu/poker"
	"log"
	"os"
	"regnivon/solv"
	"runtime/pprof"
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


	oopHands := "KQs, QJs, JTs, T9s, 98s, 77+"
	ipHands := "KQs, QJs, JTs, T9s, 98s, 77+"
	board := []poker.Card{poker.NewCard("Ac"), poker.NewCard("7s"), poker.NewCard("5s"),
		}//poker.NewCard("3d"), poker.NewCard("2h")



	//oopHands := "AA, KK, QQ, JJ, TT, 99, 88, 77, 66, 55, 44, 33, 22, AK, AQ, AJ, AT, A9, A8, A7, A6, A5, A4, A3, A2, KQ, KJ, KT, K9, K8, K7, K6, K5, K4, K3, K2, QJ, QT, Q9, Q8, Q7, Q6, Q5, Q4, Q3, Q2, JT, J9, J8, J7, J6, J5, J4, J3, J2, T9, T8, T7, T6, T5, T4, T3, T2, 98, 97, 96, 95, 94, 93, 92, 87, 86, 85, 84, 83, 82, 76, 75, 74, 73, 72, 65, 64, 63, 62, 54, 53, 52, 43, 42, 32"
	//ipHands := "AA, KK, QQ, JJ, TT, 99, 88, 77, 66, 55, 44, 33, 22, AK, AQ, AJ, AT, A9, A8, A7, A6, A5, A4, A3, A2, KQ, KJ, KT, K9, K8, K7, K6, K5, K4, K3, K2, QJ, QT, Q9, Q8, Q7, Q6, Q5, Q4, Q3, Q2, JT, J9, J8, J7, J6, J5, J4, J3, J2, T9, T8, T7, T6, T5, T4, T3, T2, 98, 97, 96, 95, 94, 93, 92, 87, 86, 85, 84, 83, 82, 76, 75, 74, 73, 72, 65, 64, 63, 62, 54, 53, 52, 43, 42, 32"


	//oopHands := "QQ, 99"
	//ipHands := "QQ, 99"


	oop := solv.HandsStringToHandRange(oopHands)
	ip := solv.HandsStringToHandRange(ipHands)

	oop = solv.RemoveConflicts(oop, board)
	ip = solv.RemoveConflicts(ip, board)


	params := solv.NewConstructionParams(0.75, 1.2)
	tree := solv.ConstructTree(6, 100, params, ip, oop, board)
	solv.OutputTree(tree)
	traversal := solv.NewTraversal(oop, ip)
	//the result should be -0.9 +0.9 for the suited game
	solv.Train(traversal, 200, tree)
	/*node := tree.GetNext(0).(*solv.GameNode)
	for index, hand := range ip {
		fmt.Printf("%v %v\n", hand.Hand, node.Strategy(index))
	} */
}
