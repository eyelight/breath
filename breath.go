// inspired by https://makersportal.com/blog/2020/3/27/simple-breathing-led-in-arduino
package breath

import (
	"machine"
	"math"
	"runtime"
	"sync"
	"time"
)

type Wave uint8

const (
	Triangular Wave = iota // default
	Circular               // sort of bouncy
	Gaussian               // configurable with Beta & Gamma
	Hold                   // pauses breathing in place
	Stop                   // stops the breather goroutine; LED pin goes low; Breathe() must be called to restart
)

type breather struct {
	*sync.Mutex
	pin       machine.Pin
	pwm       *machine.PWM
	pChan     uint8
	pattern   Wave
	delay     time.Duration
	smoothing uint16
	beta      float32
	gamma     float32
	step      int
	lastStep  time.Time
	holding   bool
	confCh    chan Conf
}

type Conf struct {
	Pattern   Wave          // the Wave to use
	Delay     time.Duration // the duration of each brightness level, eg 1ms
	Smoothing uint16        // the number of increments in the Pattern
	Beta      float32       // applicable to Gaussian; probably 0.5
	Gamma     float32       // applicable to Gaussian; 0.01 mostly dark, 1.0 mostly bright
}

type Breather interface {
	Breathe(Conf)
	Conf() Conf
}

// NewBreather returns a breather attached to a given (pre-configured) pin
func New(pin machine.Pin, pwm *machine.PWM) Breather {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pwm.Configure(machine.PWMConfig{})
	b, err := pwm.Channel(pin)
	if err != nil {
		println("could not obtain channel for pin")
	}
	return &breather{pin: pin, pwm: pwm, pChan: b}
}

// Conf returns the breather's currently loaded config as a BreatherConf
func (b *breather) Conf() Conf {
	return Conf{
		Pattern:   b.pattern,
		Smoothing: b.smoothing,
		Delay:     b.delay,
		Beta:      b.beta,
		Gamma:     b.gamma,
	}
}

// Breathe starts a breather and/or changes its cadence;
// to stop a breather, send a Conf
func (b *breather) Breathe(conf Conf) {
	switch conf.Pattern {
	case Stop:
		if b.confCh != nil {
			b.Lock()
			close(b.confCh)
			b.Unlock()
		}
	default:
		if b.confCh != nil {
			b.confCh <- conf
		} else {
			b.confCh = make(chan Conf, 1)
			go b.breathe()
			b.confCh <- conf
		}
	}
}

// breathe does the breathing in a goroutine
func (b *breather) breathe() {
	for {
		select {
		case cnf, ok := <-b.confCh:
			if !ok { // channel closure makes us clean up & exit
				b.pin.Low()
				b.Lock()
				if b.confCh != nil {
					b.confCh = nil
				}
				b.Unlock()
				runtime.GC()
				runtime.Goexit()
			} else {
				if b.holding && cnf.Pattern != Hold {
					b.holding = false
				}
				b.pattern = cnf.Pattern
				b.delay = cnf.Delay
				b.smoothing = cnf.Smoothing
				b.beta = cnf.Beta
				b.gamma = cnf.Gamma
			}
		default:
			// do the things according to config
			if time.Since(b.lastStep) >= b.delay {
				switch b.pattern {
				case Triangular:
					b.stepTriangular()
				case Circular:
					b.stepCircular()
				case Gaussian:
					b.stepGaussian()
				case Hold:
					b.holding = true
				case Stop:
					// should be unreachable
				}
			} else {
				runtime.Gosched()
			}
		}
	}
}

func (b *breather) stepTriangular() {
	if !b.holding {
		b.lastStep = time.Now()
		b.pwm.Set(b.pChan, (b.pwm.Top() * (1.0 - (2*(abs(b.step)/b.pwm.Top()) - 1.0))))
		b.incrementStep()
	}
}

func (b *breather) stepCircular() {
	if !b.holding {
		b.lastStep = time.Now()
		b.pwm.Set(b.pChan, uint32((float64(b.pwm.Top()) * (math.Sqrt(1.0 - math.Pow(math.Abs((2.0*(float64(b.step)/float64(b.smoothing)))-1.0), 2.0))))))
		b.incrementStep()
	}
}

func (b *breather) stepGaussian() {
	if !b.holding {
		b.lastStep = time.Now()
		b.pwm.Set(b.pChan, uint32(float64(b.pwm.Top())*(math.Exp(-(math.Pow(((float64(b.step)/float64(b.smoothing))-float64(b.beta))/float64(b.gamma), 2.0))/2.0))))
		b.incrementStep()
	}
}

func (b *breather) incrementStep() {
	b.step++
	if b.step == int(b.pwm.Top()) {
		b.step = -b.step
	}
}

func abs(i int) uint32 {
	if i < 0 {
		return uint32(-i)
	}
	return uint32(i)
}
