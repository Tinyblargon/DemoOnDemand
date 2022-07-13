package concurrency

func DecideMinimumTreads(numberOfObjects, concurrency uint) uint {
	if concurrency == 0 {
		concurrency = 1
	} else if numberOfObjects < concurrency {
		concurrency = numberOfObjects
	}
	return concurrency
}

func ChannelLooperError(ret chan error, cycles uint) (err error) {
	counter := 0
	for e := range ret {
		counter++
		if e != nil {
			err = e
			break
		}
		if counter == int(cycles) {
			break
		}
	}
	return
}
