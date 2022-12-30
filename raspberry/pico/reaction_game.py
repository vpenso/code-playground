import machine
import utime
import urandom

pressed = False
fastest_button = None

led = machine.Pin(15, machine.Pin.OUT)
left_button = machine.Pin(14, machine.Pin.IN, machine.Pin.PULL_DOWN)
right_button = machine.Pin(16, machine.Pin.IN, machine.Pin.PULL_DOWN)

# callback function for the interrupt
def button_handler(pin):
    global pressed
    if not pressed:
        pressed = True
        global fastest_button
        fastest_button = pin
        
# set up the interrupts
left_button.irq(trigger=machine.Pin.IRQ_RISING, handler=button_handler)
right_button.irq(trigger=machine.Pin.IRQ_RISING, handler=button_handler)

led.on()
utime.sleep(urandom.uniform(5, 10))
led.off()

while fastest_button is None:
    utime.sleep(1)

if fastest_button is left_button:
    print("Left player wins...")
elif fastest_button is right_button:
    print("Right player wins...")
