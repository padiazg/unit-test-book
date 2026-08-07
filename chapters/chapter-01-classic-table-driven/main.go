package classic_table_driven

// CRAPEntry represents a function with its computed CRAP score.
type CRAPEntry struct {
	FuncName      string
	EffectiveCRAP float64
}

// Entries wraps a list of CRAP entries.
type Entries struct {
	List []CRAPEntry
}

// ThresholdExceeded returns true if any entry exceeds the given threshold.
func (e *Entries) ThresholdExceeded(threshold float64) bool {
	for _, entry := range e.List {
		if entry.EffectiveCRAP > threshold {
			return true
		}
	}
	return false
}
