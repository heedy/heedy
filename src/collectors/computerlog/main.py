import sys
from PyQt4 import QtGui,QtCore, uic

import log

import logging

logging.basicConfig(level=logging.DEBUG,filename="cache.log")

class OptionsWindow(QtGui.QMainWindow):
    def __init__(self,windowIcon,l,parent=None):
        super(OptionsWindow,self).__init__(parent=parent)

        self.l = l

        uic.loadUi("optionswindow.ui",self)

        self.setWindowTitle("ConnectorDB Logger Options")
        self.setWindowIcon(windowIcon)

        self.saveButton.clicked.connect(self.saveClicked)

        self.load()

    def load(self):
        #Now load the current values of the settings
        self.deviceName.setText(self.l.cache.name)
        self.apiKey.setText(self.l.cache.apikey)
        self.serverUrl.setText(self.l.cache.url)

        d = self.l.cache.data

        if not d["keypresses"]:
            self.log_keypresses.setCheckState(QtCore.Qt.Unchecked)
        if not d["activewindow"]:
            self.log_activewindow.setCheckState(QtCore.Qt.Unchecked)

        self.datapointFrequency.setValue(d["gathertime"]/60.)

        self.syncFrequency.setValue(self.l.cache.syncperiod/60.)

        self.printOptions()

    def printOptions(self):
        logging.info("------------------")
        logging.info("Option Values:")
        logging.info("Device Name: %s",self.deviceName.text())
        logging.info("API KEY: %s",self.apiKey.text())
        logging.info("Server URL: %s",self.serverUrl.text())
        logging.info("Keypresses: %i",bool(self.log_keypresses.checkState()))
        logging.info("ActiveWindow: %i",bool(self.log_activewindow.checkState()))
        logging.info("GatherTime: %f",self.datapointFrequency.value())
        logging.info("Sync Time: %f",self.syncFrequency.value())
        logging.info("------------------")

    def saveClicked(self):
        self.printOptions()

        self.l.cache.name = str(self.deviceName.text())
        self.l.cache.apikey = str(self.apiKey.text())
        self.l.cache.url = str(self.serverUrl.text())
        self.l.cache.syncperiod = float(self.syncFrequency.value())*60

        d = self.l.cache.data
        d["keypresses"] = bool(self.log_keypresses.checkState())
        d["activewindow"] = bool(self.log_activewindow.checkState())
        d["gathertime"] = float(self.datapointFrequency.value())*60
        self.l.cache.data = d

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

        self.l = log.DataCache()

        self.optionsWindow = OptionsWindow(self.icon,self.l)


        if self.l.cache.data["isrunning"]:
            toggleAction.setChecked(QtCore.Qt.Checked)

    def exitButtonPressed(self):
        self.l.stop()
        sys.exit(0)
    
    def optionsButtonPressed(self):
        self.optionsWindow.show()

    def logToggleButtonPressed(self):

        if self.toggleAction.isChecked():
            err = self.l.setupstreams()
            if len(err) > 0:
                self.toggleAction.setChecked(QtCore.Qt.Unchecked)
                self.optionsWindow.show()
                QtGui.QMessageBox.warning(self.optionsWindow,"Error",err)
            else:
                val = self.l.start()
                if not val:
                    self.toggleAction.setChecked(QtCore.Qt.Unchecked)
                    self.optionsWindow.show()
                    QtGui.QMessageBox.warning(self.optionsWindow,"Error","There was an error connecting to the server... are the settings correct?")
        else:
            self.l.stop()
            d = self.l.cache.data
            d["isrunning"]=False
            self.l.cache.data = d


if __name__=="__main__":

    app = QtGui.QApplication(sys.argv)
    mainIcon = QtGui.QIcon("ic_launcher.png")
    tray = MainTray(mainIcon,app)
    tray.show()
    sys.exit(app.exec_())
