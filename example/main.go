package main

import (
	"machine"
	"time"

	"github.com/eyelight/breath"
)

func main() {
	p := machine.PWM0
	p.Configure(machine.PWMConfig{})
	led := breath.New(machine.LED, p)

	bouncy := breath.Conf{
		Pattern:   breath.Circular,
		Smoothing: 750,
		Delay:     1 * time.Millisecond,
	}

	pingy := breath.Conf{
		Pattern:   breath.Gaussian,
		Smoothing: 750,
		Delay:     1 * time.Millisecond,
		Beta:      0.5,
		Gamma:     0.01,
	}

	relaxy := breath.Conf{
		Pattern:   breath.Gaussian,
		Smoothing: 1250,
		Delay:     1 * time.Millisecond,
		Beta:      0.5,
		Gamma:     0.15,
	}

	for {
		println("Breathing bouncy")
		led.Breathe(bouncy)
		println(led.Conf())
		time.Sleep(time.Second * 5)

		println("Breathing pingy")
		led.Breathe(pingy)
		println(led.Conf())
		time.Sleep(time.Second * 5)

		println("Breathing relaxy")
		led.Breathe(relaxy)
		println(led.Conf())
		time.Sleep(time.Second * 5)
	}
}
