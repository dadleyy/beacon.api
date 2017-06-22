package bg

import "sync"

// Processor is an interface that defines a background-task with async safeguards
type Processor interface {
	Start(*sync.WaitGroup, KillSwitch)
}
