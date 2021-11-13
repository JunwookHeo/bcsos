package config

// Time period for Remove data
// T0SC0 :  After the time, no access data will be removed from local storage
const T0SC0 int = 60 * 2 // second

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
