import machine
import utime

potentiometer = machine.ADC(26)
led = machine.PWM(machine.Pin(15))
led.freq(1000)

while True:
    print(potentiometer.read_u16())
    led.duty_u16(potentiometer.read_u16())
    utime.sleep(1)