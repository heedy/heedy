#The keypress and window text logger for windows

import win32api
import win32gui
import win32console
import pyHook,pythoncom
import time

from multiprocessing import Process,Value

def runKeylogger(val):
    def OnKeyboardEvent(event):
        val.value +=1

    hm = pyHook.HookManager()
    hm.KeyDown = OnKeyboardEvent
    hm.HookKeyboard()

    pythoncom.PumpMessages()

class DataGatherer(object):
    def __init__(self):

        #Hide the console window if it is visible
        win=win32console.GetConsoleWindow()
        win32gui.ShowWindow(win,0)

        #Have the "activation variables defined for all logging activities, but they only
        #enable/disable the background processes on start and stop
        self.__log_keypresses = True
        self.log_activewindow = True


        self.keypress_number = Value('i',0)
        self.keylogger_process = None

        self.running=False

    @property
    def log_keypresses(self):
        return self.__log_keypresses
    @log_keypresses.setter
    def log_keypresses(self,val):
        if self.running and val==True:
            self.start_keylogger()
        elif val==False:
            self.stop_keylogger()

    def keypresses(self):
        v = self.keypress_number.value
        self.keypress_number.value=0
        return v
    def windowtext(self):
        return win32gui.GetWindowText(win32gui.GetForegroundWindow())

    def start(self):
        self.start_keylogger()
        self.running=True

    def start_keylogger(self):
        if self.keylogger_process is None and self.log_keypresses:
            self.keylogger_process = Process(target=runKeylogger,args=(self.keypress_number,))
            self.keylogger_process.daemon = True
            self.keylogger_process.start()

    def stop(self):
        self.stop_keylogger()
            
        self.running=False


    def stop_keylogger(self):
        #Internal - not exposed
        if self.keylogger_process is not None:
            self.keylogger_process.terminate()
            self.keylogger_process = None
            self.keypress_number.value = 0
