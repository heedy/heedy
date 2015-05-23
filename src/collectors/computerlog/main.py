import sys
from PyQt4 import QtGui,QtCore

class OptionsWindow(QtGui.QWidget):
    def __init__(self,windowIcon,parent=None):
        super(OptionsWindow,self).__init__(parent=parent)

        self.resize(400,400)
        self.setWindowTitle("ConnectorDB Logger Options")
        self.setWindowIcon(windowIcon)

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


    def exitButtonPressed(self):
        sys.exit(0)
    def optionsButtonPressed(self):
        self.optionsWindow.show()

    def logToggleButtonPressed(self):
        print "TOGGLE LOGGING:",self.toggleAction.isChecked()
          

if __name__=="__main__":
    
    app = QtGui.QApplication(sys.argv)
    mainIcon = QtGui.QIcon("ic_launcher.png")
    tray = MainTray(mainIcon,app)
    tray.show()
    sys.exit(app.exec_())