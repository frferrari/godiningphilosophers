package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

const maxPhilosophers = 5 // There are five philosophers around the table
const maxChopSticks = 5   // There are five chopticks on the table
const maxTimeToEat = 3    // philosophers can eat max 3 times

// ChopStick represents a chopstick along with a meachnisme to lock it
type ChopStick struct{ sync.Mutex }

// Philosopher allows to handle the process of eating for a philosopher, he has :
// - a unique identifier (from 0 to maxPhilosophers)
// - a count of how many times he has been eating (he should not eat more than maxTimeToEat)
// - access to 2 chopsticks,
// - and a channel in which the Host sends a message to allow/deny the philosopher to eat
type Philosopher struct {
	id                            int
	countEating                   int
	leftChopStick, rightChopStick *ChopStick
	feedbackChannel               chan bool
}

// Request is used by the philosophers to send messages to the Host :
// - wantToEat when they would like to eat, this can be accepted or rejected by the Host
// - finishedEating when a philosopher wants to signal that he has finished eating
type Request struct {
	command     string
	philosopher Philosopher
}

// Below are the allowed command for the Request struct
const wantToEat = "wantToEat"
const finishedEating = "finishedEating"

// eat function allows to start the process of eating for a philosopher
// To eat a philosopher sends a request to the Host, who can accept or reject the request
// - if the request to eat is accepted by the Host through the philosopher's feedback channel, the philosopher :
//   * locks the chopstick he has access to
//   * then eats during some time
//   * unlocks the chopsticks
//   * increments his count of eating
//   * and sends a message to the Host that he has finished eating
// This process loops until the philosopher reaches 3 times eating, at which point the process stops
func (philosopher Philosopher) eat(requestChan chan Request, wg *sync.WaitGroup) {
	philosopher.countEating = 0

	for philosopher.countEating < 3 {
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)

		requestChan <- Request{command: wantToEat, philosopher: philosopher}
		isPhilosopherAllowedToEat := <-philosopher.feedbackChannel

		if isPhilosopherAllowedToEat {
			philosopher.leftChopStick.Lock()
			philosopher.rightChopStick.Lock()
			fmt.Printf("starting  eating %d (%d)\n", philosopher.id, philosopher.countEating)
			time.Sleep(time.Duration((rand.Intn(500) + 50)) * time.Millisecond)
			fmt.Printf("finishing eating %d (%d)\n", philosopher.id, philosopher.countEating)
			philosopher.rightChopStick.Unlock()
			philosopher.leftChopStick.Unlock()

			philosopher.countEating++

			wg.Done()

			requestChan <- Request{command: finishedEating, philosopher: philosopher}
		}
	}

	close(philosopher.feedbackChannel)
}

// Start of the program
func main() {
	// Creating the ChopSticks
	var chopSticks = make([]*ChopStick, maxChopSticks)
	for chopStick := 0; chopStick < maxChopSticks; chopStick++ {
		chopSticks[chopStick] = new(ChopStick)
	}

	// Creating the Philosophers
	var philosophers = make([]*Philosopher, maxPhilosophers)
	for philosopher := 0; philosopher < maxPhilosophers; philosopher++ {
		// philosopher 0 will have chopstick 0 and 1
		// philosopher 1 will have chopstick 1 and 2
		// philosopher 2 will have chopstick 2 and 3
		// philosopher 3 will have chopstick 3 and 4
		// philosopher 4 will have chopstick 4 and 0
		var leftChopStickID = philosopher
		var rightChopStickID = (philosopher + 1) % maxPhilosophers
		philosophers[philosopher] = &Philosopher{
			id:              philosopher,
			countEating:     0,
			leftChopStick:   chopSticks[leftChopStickID],
			rightChopStick:  chopSticks[rightChopStickID],
			feedbackChannel: make(chan bool)}
	}

	// A wait group to allow the main program to wait for all the philosophers to eat 3 times
	var wg sync.WaitGroup
	wg.Add(maxPhilosophers * maxTimeToEat)

	// A channel in which the philosophers send their requests to the Host
	var requestChan = make(chan Request)

	// The host will ensure that a max of 2 philosophers eat at the same time
	// and that this philosophers are not neighborhood otherwise we could
	// end up with a deadlock
	go Host(requestChan)

	// Create and start the goroutines for the philosophers
	for _, philosopher := range philosophers {
		go philosopher.eat(requestChan, &wg)
	}

	// Wait for all the philosophers to eat 3 times
	wg.Wait()

	close(requestChan)

	fmt.Println("All philosophers have finished eating, good bye")
}

// Host receives requests to eat from the philosophers, the host decide to accept or reject each request and ensures that :
// - only 2 philosophers eat at the same time
// - the 2 philosophers eating at the same time cannot be neighborhood
// The Host also processes the messages sent by the philosophers when they have finished eating, this allows the Host
//   to authorize only 2 philosophers to eat at the same time
func Host(requestChan chan Request) {
	var philosophersEating = make(map[int]Philosopher)

	for {
		request := <-requestChan

		switch request.command {
		case wantToEat:
			if len(philosophersEating) == 0 {
				philosophersEating[request.philosopher.id] = request.philosopher
				AcceptRequestToEat(&request.philosopher)
			} else if len(philosophersEating) == 1 {
				var keys []int
				for k := range philosophersEating {
					keys = append(keys, k)
				}
				var philosopherCurrentlyEating = keys[0]
				var philosopherAskingToEat = request.philosopher.id
				// Neighborhoods ?
				if philosopherCurrentlyEating == 0 && philosopherAskingToEat == maxPhilosophers {
					RejectRequestToEat(&request.philosopher, "Neighborhood 0-")
				} else if philosopherCurrentlyEating == maxPhilosophers && philosopherAskingToEat == 0 {
					RejectRequestToEat(&request.philosopher, "Neiborhood -0")
				} else if math.Abs(float64(philosopherAskingToEat-philosopherCurrentlyEating)) == 1.0 {
					RejectRequestToEat(&request.philosopher, "Neighborhood")
				} else if philosopherAskingToEat == philosopherCurrentlyEating {
					RejectRequestToEat(&request.philosopher, "Philosopher alread eating")
				} else {
					AcceptRequestToEat(&request.philosopher)
				}
			} else {
				RejectRequestToEat(&request.philosopher, "All allowed philoshopers are already eating")
			}
		case finishedEating:
			delete(philosophersEating, request.philosopher.id)
		}
	}
}

// RejectRequestToEat sends a message back to the philosopher denying him to eat
func RejectRequestToEat(philosopher *Philosopher, rejectReason string) {
	fmt.Printf("Host rejects request to eat from %d, reason %s\n", philosopher.id, rejectReason)
	philosopher.feedbackChannel <- false
}

// AcceptRequestToEat sends a message back to the philosopher allowing him to eat
func AcceptRequestToEat(philosopher *Philosopher) {
	fmt.Printf("Host accepts request to eat from %d\n", philosopher.id)
	philosopher.feedbackChannel <- true
}
