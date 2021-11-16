package config

// Block create period
const BLOCK_CREATE_PERIOD int = 5

// The number of transaction in a block
// IF NUM_TRANSACTION_BLOCK == 0, choose random between 3 to 6
const NUM_TRANSACTION_BLOCK int = 6

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
/////////////////////////////////////////////////
// P	Lambda	Time
// 0.5	0.1		6.931471806
// 0.6	0.1		9.162907319
// 0.7	0.1		12.03972804
// 0.8	0.1		16.09437912
// 0.9	0.1		23.02585093

// The basic unit of time (T)
const BASIC_UNIT_TIME int = 60 // 60 seconds
const PROBABILITY_FACTOR_50 float32 = 0.69
const PROBABILITY_FACTOR_70 float32 = 1.20
const PROBABILITY_FACTOR_90 float32 = 2.30

// Time period for Remove data : T x TSC0
// TSC0 :  After the time, no access data will be removed from local storage.
// 1 Minuite : The start time for TS20 to remove objects is 23 minuites after starting simulation
// so about 50 minute is needed for total simulation time
// IoT+normal_fridge_1.log has 40057 transactions. We generate a block with 6 transactions
// and create every 5 seconds, so total simulation time will be around 9.3 hours.
// Thus, 10 minuites is good for 500 minuites(8.3 hours)
const RATE_TSC int = 10 // x BASIC_UNIT_TIME

// Lambda for Exponential Distribution
// the number of event in TSC0
const LAMBDA_ED float32 = 0.1

// TSC0 = RATE_TSC x BASIC_UNIT_TIME * (PROBABILITY_FACTOR / LAMBDA_ED)
const TSC0I int = int(float32(RATE_TSC*BASIC_UNIT_TIME) * (float32(PROBABILITY_FACTOR_50) / LAMBDA_ED))
const TSC1I int = int(float32(RATE_TSC*BASIC_UNIT_TIME) * (float32(PROBABILITY_FACTOR_70) / LAMBDA_ED))
const TSC2I int = int(float32(RATE_TSC*BASIC_UNIT_TIME) * (float32(PROBABILITY_FACTOR_90) / LAMBDA_ED))

// Array for TSC0 ~ 3
var TSCX = [...]int{TSC0I, TSC1I, TSC2I, 0}

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
