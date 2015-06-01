

class DataGatherer(object):
    def __init__(self):
        self.log_activewindow = True
        self.log_keypresses = True

    def keypresses(self):
        #Gets the number of keypresses since last call of this function
        pass

    def windowtext(self):
        #Gets the titlebar text of currently active window
        pass
    
    def start(self):
        #Starts any background processes needed to get it working
        pass
    def stop(self):
        #stops logging background processes
        pass