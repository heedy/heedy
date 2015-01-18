import argparse
import ConfigParser
import os
import logging
import logging.handlers

logger = logging.getLogger("Setup")

class Configuration(object):
    """
    This class holds the configuration of the program. It loads every parameter both from files
    and from the command line
    """

    cfg = None
    console_logformat = "%(name)s:%(levelname)s - %(message)s"
    file_logformat = "%(levelname)s:%(name)s:%(created)f - %(message)s"

    def __init__(self,args = {},description=""):

        #Run setup only if the config is not None
        if (Configuration.cfg is None):
            self.initLogger()
            
            self.initCfg(description,args)
            
            #If there is a log file set up, write to it!
            if ("logfile" in self.cfg and self.cfg["logfile"] is not None):
                self.logfile(self.cfg["logfile"])

            logger.info("CONFIGURATION: %s",str(self.cfg))

    def initLogger(self):
        logger = logging.getLogger()
        logger.setLevel(logging.INFO)
        logger.propagate= False

        #Write log messages to console
        ch = logging.StreamHandler()
        ch.setLevel(logging.INFO)
        ch.setFormatter(logging.Formatter(Configuration.console_logformat))
        logger.addHandler(ch)
        
    def logfile(self,fname):
        logger.info("Writing to log '%s'",fname)
        #Add a log file to which to log
        ch = logging.handlers.RotatingFileHandler(fname,maxBytes = 1024*1024*10,backupCount=5)
        ch.setLevel(logging.INFO)
        ch.setFormatter(logging.Formatter(Configuration.file_logformat))
        
        logging.getLogger().addHandler(ch)
        

    def initCfg(self,description,args):
        self.cfg = {"name": "default"}

        #Add default values to config options
        for arg in args:
            if ("default" in args[arg]):
                self.cfg[arg] = args[arg]["default"]


        parser = argparse.ArgumentParser(description=description)

        for arg in args:
            a  = {"help": args[arg]["help"]}
            if ("type" in args[arg]):
                a["type"] = args[arg]["type"]
            sec = None
            if ("s" in args[arg]):
                sec = "-"+args[arg]["s"]
            arg= "--"+arg

            if (sec is not None):
                parser.add_argument(sec,arg,**a)
            else:
                parser.add_argument(arg,**a)

        #Now load the default stuff that gets applied each time

        #The configuration file to load.
        parser.add_argument("-c","--config",help="The configuration file to use")

        #The following options override values within the configuration file
        parser.add_argument("-l","--logfile",help="The log file location")

        parser.add_argument("name",nargs="?",help="The name of config to load")

        #Parse the command line
        cmdline = vars(parser.parse_args())

        if (cmdline["config"] is not None):
            self.loadFile(cmdline["config"],cmdline["name"])
         

        #Now load the cmdline args to overload any config file settings
        for key in cmdline:
            if (cmdline[key] is not None and key!="name"):
                self.cfg[key] = cmdline[key]

        #Finally, make sure all args are of the correct type
        for arg in args:
            if ("type" in args[arg] and arg in self.cfg):
                self.cfg[arg] = args[arg]["type"](self.cfg[arg])

    def loadFile(self,fname,lname=None):
        logger.info("Loading config '%s'",fname)
        if not (os.path.exists(fname)):
            raise Exception("Could not find configuration file")

        #Reads configuration file
        config = ConfigParser.SafeConfigParser()
        config.read(fname)

        def loadSettings(name):
            if (name!="default" and not config.has_section(name)):
                return

            for option in config.options(name):
                self.cfg[option] = config.get(name,option)

        loadSettings("default")

        oldname = "default"
        while (oldname != self.cfg["name"]):
            oldname = self.cfg["name"]
            loadSettings(self.cfg["name"])

        #Now, since default is loaded, we check if lname is set, and load config from there
        if (lname is not None):
            self.cfg["name"] = lname
            while (oldname != self.cfg["name"]):
                oldname = self.cfg["name"]
                loadSettings(self.cfg["name"])

    def __str__(self):
        return str(self.cfg)

    def __getitem__(self,i):
        if i in self.cfg:
            return self.cfg[i]
        return None
    def keys(self):
        return self.cfg.keys()


if (__name__=="__main__"):
    l = logging.getLogger("hello")

    l.error("Here is 1")

    print Configuration()

    l.error("Here is 2")