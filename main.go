package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	customerIdCounter = 1
	barberIdCounter   = 1
)

// barber
type Barber struct {
	ID   int
	Name string
	lock sync.Mutex
}

// customer
type Customer struct {
	ID   int
	Name string
}

func main() {
	customerStopCh := make(chan struct{})
	barberStopCh := make(chan struct{})
	shopTimer := time.NewTimer(time.Duration(15 * time.Second))

	waitingList := make(chan Customer, 3)

	barberWaitGrp := sync.WaitGroup{}
	customerWaitGrp := sync.WaitGroup{}

	// loop for the barber
	for i := 0; i < 2; i++ {
		barberWaitGrp.Add(1)
		//new barber
		newBarber := Barber{
			ID:   barberIdCounter,
			Name: "barber" + strconv.Itoa(barberIdCounter),
			lock: sync.Mutex{},
		}

		// call the barber func as go routine
		go BarberFunc(&newBarber, waitingList, &barberWaitGrp, &customerWaitGrp, barberStopCh)
		barberIdCounter++
	}

	// for the customer
	go func(stopCh <-chan struct{}) {
		// open the shop and start accepting customers
		shopTimer = time.NewTimer(time.Duration(15 * time.Second))
		isShopOpen := true
		for isShopOpen {
			select {
			case <-stopCh:
				fmt.Println("Shop is closed, no more customers will be accepted !!")
				isShopOpen = false // shop is closed
				break
			default:
				customerWaitGrp.Add(1)
				// create new customer
				newCustomer := Customer{
					ID:   customerIdCounter,
					Name: "customer" + strconv.Itoa(customerIdCounter),
				}

				waitingList <- newCustomer
				customerIdCounter++
			}
			time.Sleep(1 * time.Second)
		}
	}(customerStopCh)

	// wait for the timer to expire
	<-shopTimer.C

	// then close channel
	close(customerStopCh)
	close(barberStopCh)

	customerWaitGrp.Wait()
	barberWaitGrp.Wait()
	fmt.Println("End of day...............")
}

// Accepting customers

// Barber doing her thing
func BarberFunc(barber *Barber, waitingList chan Customer, wg, cWg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()

	// barber starts working
	barberWorking := true
	for barberWorking {
		barber.lock.Lock()
		select {
		case customer := <-waitingList: // customer exist

			fmt.Println(barber.Name + " starts cutting hair of " + customer.Name)
			time.Sleep(2 * time.Second) // cuts hair
			fmt.Println("Haircut done for " + customer.Name)
			cWg.Done() // mark the customer as done
			break
		case <-stopCh:
			fmt.Println(barber.Name + " sign off for the day :)")
			barberWorking = false
		default:
			fmt.Println(barber.Name + " is sleeping !!")
			time.Sleep(2 * time.Second)
		}
		barber.lock.Unlock()
	}
}
