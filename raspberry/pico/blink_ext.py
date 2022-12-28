 
import machine
import utime
 
# GP15 ...physical port 20
led_external = machine.Pin(15, machine.Pin.OUT)
 
while True:
    led_external.toggle()
    utime.sleep(5)
