import machine
import utime
 
# GP15 ...physical pin 20
led_external = machine.Pin(15, machine.Pin.OUT)
# GP14 ...physical pin 19
button = machine.Pin(14, machine.Pin.IN, machine.Pin.PULL_DOWN)
 
while True:
    if button.value() == 1:
        led_external.value(1)
        utime.sleep(1)
    led_external.value(0)
