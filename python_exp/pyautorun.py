import pyautogui as pg
import time

size = pg.size()
print(size)

button5location = pg.locateOnScreen('./common/resources/xterm.png') 
print(button5location)

time.sleep(3)
button5location = pg.locateOnScreen('./common/resources/vscode.png') 
print(button5location)

image = pg.screenshot()
image.save('testing.png')
