# breath
Let your LEDs breathe

## About
Using TinyGo on microcontroller targets, you can use PWM to fade your LEDs. This package lets you configure different "breathing" patterns such as Triangular, Circular, and Gaussian. This package is inspired by [this blog post](https://makersportal.com/blog/2020/3/27/simple-breathing-led-in-arduino).

## Usage
First, set up your desired PWM peripheral.
```golang
p := machine.PWM0
p.Configure(machine.PWMConfig{})
```
Next, call `New` to turn your LED into a `Breather`. Pass your desired pin along with the PWM peripheral you just created.
```golang
led := breath.New(machine.LED, p)
```
At this point, your LED is not breathing, but it is ready to receive a `breath.Conf` via the `Breathe` method, which passes a `breath.Conf` to a long-lived goroutine to which you can modify behavior as the situation changes by calling Breathe again with a new `breath.Conf`. 

If you want to create multiple breath configurations, you can easily pass them by name. Let's create some.

```golang
bouncy := breath.Conf{
	Pattern:   breath.Circular,
	Smoothing: 750,
	Delay:     1 * time.Millisecond,
}

relaxy := breath.Conf{
	Pattern:   breath.Gaussian, // which wave to use
	Smoothing: 1024, // number of steps in wave
	Delay:     1 * time.Millisecond, // duration per step
	Beta:      0.5, // boring input to the gaussian function
	Gamma:     0.21, // exciting input to the gaussian function
}

hold := breath.Conf{
	Pattern: breath.Hold,
}

stop := breath.Conf{
	Pattern: breath.Stop,
}
```

We're ready to start breathing. The first time you call `Breathe`, a goroutine is spawned, which is intended to be a good citizen by calling the garbage collector & scheduler at appropriate times. Let's start your Breather:

```golang
led.Breathe(relaxy)
```

You can make the breather 'hold' its breath without exiting the goroutine by passing `Pattern: Hold` in a call to `Breathe`. The led will remain paused at the current PWM value, and a subsequent call with an active pattern will resume where the breathing left off. 

```golang
led.Breathe(hold)
```

The breather can be stopped completely by passing a `breath.Conf` with a `Pattern: Stop` in a subsequent call to `Breathe`. This will cause the goroutine to exit and will bring the attached pin low. 


```golang
led.Breathe(stop)
```

The breather is still around even though the goroutine has stopped. To start it again, merely call `Breathe` again with a new configuration. The same breather will spawn a new goroutine, the old one having cleaned up its resources. And yes, you can change the config on the fly without holding or stopping.

```golang
led.Breathe(bouncy)
time.Sleep(5 * time.Second)
led.Breathe(relaxy)
```

If you ever get curious about the current `breath.Conf`, you can call `Conf`, which is the only other method in the API. `Conf` will return the currently loaded values for your breather.

```golang
if led.Conf().Pattern != breath.Stop {
	led.Breathe(stop)
}
```