package main

import (
	"fmt"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"time"
)

var maxPowerMessure float64

func messurePower(outgoingChannel chan<- int8)  {

	//Initial values
	maxPowerMessure = 0

	var err error
	a := raspi.NewAdaptor()
	ads1015 := i2c.NewADS1015Driver(a)
	// Adjust the gain to be able to read values of at least 5V
	ads1015.DefaultGain, err = ads1015.BestGainForVoltage(5.0)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("ads1015.DefaultGain", ads1015.DefaultGain)
	}

	ads1015.WithAddress(72)

	/*
		analogRead, err := ads1015.AnalogRead("0-1")
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("analogRead", analogRead)
		}
	*/

	work := func() {
		gobot.Every(100*time.Millisecond, func() {
			v, err := ads1015.ReadWithDefaults(0)

			if err != nil {
				log.Fatal(err)
			} else {
				// adjust maxPowermessure if greater value found
				if v > maxPowerMessure {
					maxPowerMessure = v
				}
				// Calculate percent vlue for power
				percent := int8(v/maxPowerMessure *100)
				fmt.Println("A0", v)
				fmt.Println("Persent", percent)

				// Write back to channel
				outgoingChannel <- percent

			}

		})
	}

	robot := gobot.NewRobot("ads1015bot",
		[]gobot.Connection{a},
		[]gobot.Device{ads1015},
		work,
	)

	robot.Start()
}

