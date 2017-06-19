package bg

import "sync"

type Processor interface {
	Start(*sync.WaitGroup)
}
