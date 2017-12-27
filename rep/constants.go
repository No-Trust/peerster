package rep

/*
    Constatns
*/

// General
const MIN_REP   float32 = 0
const MAX_REP   float32 = 1
const INIT_REP  float32 = 0.5

const REP_RANGE float32 = MAX_REP - MIN_REP

// Signature-based reputation
const SIG_INCREASE_LIMIT float32 = 0.1
const SIG_DECREASE_LIMIT float32 = 0.8

// Contribution-based reputation
const CONTRIB_ALPHA           float32 = 0.4
const CONTRIB_ONE_MINUS_ALPHA float32 = 1 - CONTRIB_ALPHA
