import sys
from PyQt4 import QtGui,QtCore, uic

import log

class OptionsWindow(QtGui.QMainWindow):
    def __init__(self,windowIcon,parent=None):
        super(OptionsWindow,self).__init__(parent=parent)
        
        uic.loadUi("optionswindow.ui",self)

        self.setWindowTitle("ConnectorDB Logger Options")
        self.setWindowIcon(windowIcon)

        self.serverUrl.addItem("https://connectordb.com")

        self.saveButton.clicked.connect(self.saveClicked)

    def saveClicked(self):
        #self.serverUrl.
        print "Options Updated"
        print "Device Name:",self.deviceName.text()
        print "API KEY:",self.apiKey.text()
        print "Server URL: ",str(self.serverUrl.currentText())
        print "Keypresses:",bool(self.log_keypresses.checkState())
        print "ActiveWindow:",bool(self.log_activewindow.checkState())
        print "GatherTime:",self.datapointFrequency.value()
        print "Sync Time:",self.syncFrequency.value()
        self.hide()

    def keyPressEvent(self, e):
        if e.key() == QtCore.Qt.Key_Escape:
            self.hide()
    
    def closeEvent(self,event):
        event.ignore()
        self.hide()
    
class MainTray(QtGui.QSystemTrayIcon):
    def __init__(self,icon,parent=None):
        super(MainTray,self).__init__(icon,parent)
        
        self.icon = icon

        menu = QtGui.QMenu()
        
        toggleAction = menu.addAction("Run Logger")
        toggleAction.setCheckable(True)
        toggleAction.triggered.connect(self.logToggleButtonPressed)
        self.toggleAction = toggleAction

        optionsAction = menu.addAction("Options")
        optionsAction.triggered.connect(self.optionsButtonPressed)

        exitAction = menu.addAction("Exit")
        exitAction.triggered.connect(self.exitButtonPressed)

        self.setContextMenu(menu)
        self.menu = menu

        self.optionsWindow = OptionsWindow(self.icon)

        self.l = log.DataCache()


    def exitButtonPressed(self):
        self.l.stop()
        sys.exit(0)
    def optionsButtonPressed(self):
        self.optionsWindow.show()

    def logToggleButtonPressed(self):
        if self.toggleAction.isChecked():
            self.l.start()
        else:
            self.l.stop()
          

if __name__=="__main__":
    
    app = QtGui.QApplication(sys.argv)
    mainIcon = QtGui.QIcon("ic_launcher.png")
    tray = MainTray(mainIcon,app)
    tray.show()
    sys.exit(app.exec_())