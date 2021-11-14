package config

// Block create period
const BLOCK_CREATE_PERIOD int = 5

// The number of transaction in a block
// IF NUM_TRANSACTION_BLOCK == 0, choose random between 3 to 6
const NUM_TRANSACTION_BLOCK int = 5

const (
	RANDOM_ACCESS_PATTERN      string = "Random_Distribution"
	EXPONENTIAL_ACCESS_PATTERN string = "Exponential_Distribution"
)
const ACCESS_FREQUENCY_PATTERN string = EXPONENTIAL_ACCESS_PATTERN

// P = 1 - e^(-lambda*t)
// Example : Probability of event occure to be 50%
// T = 0.69/Lambda
// Lambda	Time
//   1       0.69
//   0.5     1.38
//   0.2     3.45
//   0.1     6.9
// The basic unit of time (T)
const BASIC_UNIT_TIME int = 60 // 60 seconds
const HALF_PROBABILITY_FACTOR float32 = 0.69

// Time period for Remove data : T x TSC0
// TSC0 :  After the time, no access data will be removed from local storage
const RATE_TSC0 int = 2 // x BASIC_UNIT_TIME

// Lambda for Exponential Distribution
// the number of event in TSC0
const LAMBDA_ED float32 = 0.1

// TSC0 = RATE_TSC0 x BASIC_UNIT_TIME * (HALF_PROBABILITY / LAMBDA_ED)
const TSC0F int = int(float32(RATE_TSC0*BASIC_UNIT_TIME) * (float32(HALF_PROBABILITY_FACTOR) / LAMBDA_ED))

//const TSC0I int = TSC0F

// Max storage class
const MAX_SC int = 3

// Max the number of peers for each storage class
const MAX_SC_PEER int = 5

// Simulator Storage class : 100
const SIM_SC int = 100

// NUM_AP_GEN/TIME_AP_GEN : The rate to access data for access pattern
// The number of transactions to be read for access pattern
const NUM_AP_GEN int = 10

// The number of time to create accessing transaction for access pattern
const TIME_AP_GEN int = 10 // Second

// The time to search neighbour nodes to update node info
const TIME_UPDATE_NEITHBOUR int = 10 //60 // Second
