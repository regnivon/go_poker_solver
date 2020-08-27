# Go Poker Solver
###Note: This project has been abandoned in favor of a c++ implementation, which is much, much faster. Thus the implementation is not maximally efficient, and te tests are somewhat broken.
 
This program solves for the nash equilibrium of flop, turn, and river subgames of Texas Hold'em poker.
It allows for extensive tree building. Results were checked against a commercial program, 
and the best response utilities calculated were the same for several given inputs. Speed was the primary issue,
and upon profiling, the GC was taking up huge amounts of time as well as a few other things like map usage in Terminal 
utility evaluations.

To run, simply use make run, or use the main in the driver folder to create your own version. Flop and turn subgames
are solved in parallel using the max number of logical cores, which can be changed by setting GOMAXPROCs. 
