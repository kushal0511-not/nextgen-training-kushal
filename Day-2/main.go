package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const N = 100

type Stock struct {
	Name string
}

type PricePoint struct {
	Price     float64
	Timestamp time.Time
	Volume    int
}

type CircularBuffer struct {
	Data       [N]PricePoint
	WriteIndex int
	Full       bool
}

func (cb *CircularBuffer) Add(pp PricePoint) {
	cb.Data[cb.WriteIndex] = pp
	cb.WriteIndex = (cb.WriteIndex + 1) % N
	if cb.WriteIndex == 0 {
		cb.Full = true
	}
}

func (cb *CircularBuffer) GetCurrentPrice() float64 {
	if !cb.Full && cb.WriteIndex == 0 {
		return 0
	}
	lastIndex := (cb.WriteIndex - 1 + N) % N
	return cb.Data[lastIndex].Price
}

func (cb *CircularBuffer) GetSMA() float64 {
	sum := 0.0
	count := cb.WriteIndex
	if cb.Full {
		count = N
	}
	if count == 0 {
		return 0
	}
	for i := 0; i < count; i++ {
		sum += cb.Data[i].Price
	}
	return sum / float64(count)
}

func (cb *CircularBuffer) GetMinMax() (float64, float64) {
	count := cb.WriteIndex
	if cb.Full {
		count = N
	}
	if count == 0 {
		return 0, 0
	}

	min := cb.Data[0].Price
	max := cb.Data[0].Price

	for i := 1; i < count; i++ {
		p := cb.Data[i].Price
		if p < min {
			min = p
		}
		if p > max {
			max = p
		}
	}
	return min, max
}

func (cb *CircularBuffer) String() string {
	curr := cb.GetCurrentPrice()
	min, max := cb.GetMinMax()
	sma := cb.GetSMA()
	return fmt.Sprintf("Price: %6.2f | Min/Max: %6.2f/%6.2f | SMA: %6.2f", curr, min, max, sma)
}

type StockExchange struct {
	Prices           map[Stock][]PricePoint
	LastNPricePoints map[Stock]*CircularBuffer
	mu               sync.RWMutex
}

func NewStockExchange() *StockExchange {
	return &StockExchange{
		Prices:           make(map[Stock][]PricePoint),
		LastNPricePoints: make(map[Stock]*CircularBuffer),
	}
}

func (se *StockExchange) updatePrice(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			se.mu.Lock()
			for stock, points := range se.Prices {
				newPP := PricePoint{
					Price:     100.0 + (rand.Float64()*20.0 - 10.0), // Variation around 100
					Timestamp: time.Now(),
					Volume:    rand.Intn(100),
				}

				// Dynamic Array Resize Logging
				oldCap := cap(points)
				se.Prices[stock] = append(points, newPP)
				newCap := cap(se.Prices[stock])
				if newCap > oldCap && oldCap > 0 {
					fmt.Printf("\r\n[LOG] Buffer resized for %s: cap %d → %d\n", stock.Name, oldCap, newCap)
				}

				// Circular Buffer update
				if _, ok := se.LastNPricePoints[stock]; !ok {
					se.LastNPricePoints[stock] = &CircularBuffer{}
				}
				se.LastNPricePoints[stock].Add(newPP)
			}
			se.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (se *StockExchange) Display() {
	se.mu.RLock()
	defer se.mu.RUnlock()

	clearScreen()
	fmt.Println("=== LIVE STOCK DASHBOARD ===")
	fmt.Printf("%-10s | %s\n", "STOCK", "ANALYTICS (Last 100)")
	fmt.Println("-----------|---------------------------------------------------------")
	for stock, cb := range se.LastNPricePoints {
		fmt.Printf("%-10s | %s\n", stock.Name, cb.String())
	}
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	se := NewStockExchange()
	stocks := []Stock{{Name: "HDFC"}, {Name: "RELIANCE"}, {Name: "TCS"}, {Name: "INFY"}, {Name: "ITC"}}

	for _, s := range stocks {
		se.Prices[s] = make([]PricePoint, 0, 10) // Start with capacity 10
		se.LastNPricePoints[s] = &CircularBuffer{}
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go se.updatePrice(ctx)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			se.Display()
		case <-ctx.Done():
			fmt.Println("\n[INFO] Shutdown signal received. Cleaning up...")
			// Small delay to show the message
			time.Sleep(500 * time.Millisecond)
			fmt.Println("[INFO] Graceful shutdown complete. Goodbye!")
			return
		}
	}
}
