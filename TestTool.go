package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

type Playing struct {
	Concurrency int `json:"concurrency"`
	PlayTimes   int `json:"play_times"`
	Bet         int `json:"bet"`
}

func main() {
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGTERM, os.Interrupt)
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer configFile.Close()

	configForPlaying, _ := ioutil.ReadAll(configFile)

	var playing Playing
	var wg sync.WaitGroup

	json.Unmarshal(configForPlaying, &playing)

	chBet := make(chan int)
	chWin := make(chan int)
	chPlayTimes := make(chan int)
	chQuit := make(chan bool)

	playingTotalBet := 0
	playingTotalWin := 0
	playingTotalPlayTimes := 0
	playingTotalRTP := 0.0

	for index := 1; index <= playing.Concurrency; index++ {
		go func() {
			defer func() {
				if err := recover(); err != nil {
					wg.Done()
				} else {
					wg.Done()
				}
			}()
			var totalBet = 0
			var totalWin = 0
			var totalPlayTimes = 0
			var rtp = 0.0
			for ; totalPlayTimes < playing.PlayTimes; totalPlayTimes++ {
				res, err := http.Get("http://localhost:3000/roll")
				if err != nil {
					log.Println(err)
				}

				rollPointRes, _ := ioutil.ReadAll(res.Body)

				rollPoint, err := strconv.ParseInt(string(rollPointRes), 10, 64)

				if err != nil {
					log.Fatal(err)
				}

				res.Body.Close()

				win := 0

				switch rollPoint {
				case 1, 5:
					win = 0
				case 2:
					win = playing.Bet * 5
				case 3:
					win = playing.Bet * 3
				case 4:
					win = playing.Bet * 8
				case 6:
					win = playing.Bet * 10
				}

				totalWin += win
				totalBet += playing.Bet

				chBet <- playing.Bet
				chWin <- win
				chPlayTimes <- 1
			}

			rtp = float64(totalWin) / float64(totalBet)
			fmt.Printf("TotalBet:%d\n", totalBet)
			fmt.Printf("TotalWin:%d\n", totalWin)
			fmt.Printf("TotalPlayTimes:%d\n", totalPlayTimes)
			fmt.Printf("RTP:%f\n", rtp)

		}()
	}

	go func() {
		for {
			select {
			case bet := <-chBet:
				playingTotalBet += bet
			case win := <-chWin:
				playingTotalWin += win
			case <-chPlayTimes:
				playingTotalPlayTimes++
			case <-chQuit:
				return
			}
		}
	}()
	go func() {
		<-chSig
		wg.Add(-playing.Concurrency)
	}()
	wg.Add(playing.Concurrency)
	wg.Wait()
	chQuit <- true
	playingTotalRTP = float64(playingTotalWin) / float64(playingTotalBet)
	fmt.Printf("PlayingTotalBet:%d \n", playingTotalBet)
	fmt.Printf("PlayingTotalWin:%d \n", playingTotalWin)
	fmt.Printf("PlayingTotalPlayTimes:%d \n", playingTotalPlayTimes)
	fmt.Printf("PlayingTotalRTP:%f \n", playingTotalRTP)
}