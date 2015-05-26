#A very simple windows keypress frequency and active window logger for connectordb

import win32api
import win32gui
import win32console
import pyHook,pythoncom
import time

from multiprocessing import Process,Value

win=win32console.GetConsoleWindow()
win32gui.ShowWindow(win,0)


def NewDatapoint(keypresses,activewindow):
    timestamp = time.time()
    datapoint = {"t": time.time(), "d": {"keypresses":keypresses,"activewindow": activewindow}}
    f=open("keylog.txt","a")
    f.write(str(datapoint)+"\n")
    f.close()


def runKeylogger(val):
    def OnKeyboardEvent(event):
        val.value +=1

    hm = pyHook.HookManager()
    hm.KeyDown = OnKeyboardEvent
    hm.HookKeyboard()

    pythoncom.PumpMessages()

if __name__=="__main__":
    num = Value('i',0)
    p = Process(target=runKeylogger,args=(num,))
    p.start()

    try:
        while True:
            time.sleep(60)
            NewDatapoint(num.value,win32gui.GetWindowText(win32gui.GetForegroundWindow()))
            num.value=0
    except:
        pass
    p.terminate()