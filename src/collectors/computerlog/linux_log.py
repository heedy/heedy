''' This file is part of ConnectorDB's computerlogger package.

Copyright 2015 Joseph Lewis <joseph@josephlewis.net>
All rights reserved

This file gathers data on common Linux systems using /dev and /proc along
with a few common commands if they are available on your system.

'''

import os
import subprocess
import re
import socket

PROCESS_SCHEMA = {}
def _get_processes():
    '''Peeks through the processes in the system and returns a list of maps
    consistent with the linux DataGatherer.get_data call

    '''

    processes = []

    return []


def _get_keypresses():
    return []

SYSINFO_SCHEMA = {
    "type":"object",
    "properties":{
        "kernel":{"type":"string", "description":"The kernel name of the OS"},
        "release":{"type":"string", "description":"The version of the kernel"},
        "processor":{"type":"string", "description":"The kind of processor you're using"},
        "os":{"type":"string", "description":"The operating system you're using"},
        "distro":{"type":"string", "description":"The specific distribution of the OS"}
    }
}
SYSINFO_TITLE = "system_information"
def _uname():
    '''Peeks through the system info and returns a list of maps
    consistent with the linux DataGatherer.get_data call

    '''

    kernel, release, proc, os, distro = "", "", "", "", ""
    # kernel name, kernel release, processor type, operating system
    # ex. Linux 3.13.0-37-generic x86_64 GNU/Linux
    try:
        results = subprocess.check_output(["uname", "-s", "-r", "-p", "-o"])
        kernel, release, proc, os = results.split()
    except subprocess.CalledProcessError:
        # we probably don't have the call
        pass

    try:
        distro = subprocess.check_output(["lsb_release", "--description"])
        distro = distro.split(":")[1].strip()
    except:
        # lsb-release isn't in all distros, nor is the description
        pass

    resultset = {
        "kernel":kernel,
        "release":release,
        "processor":proc,
        "os":os,
        "distro": distro
    }

    return [{"schema":SYSINFO_SCHEMA, "title":SYSINFO_TITLE, "value":resultset}]




IP_SCHEMA = {
            "type":"object",
            "properties":{
                "hostname":{"type":"string", "description":""},
                "ip":{"type":"string", "description":"The IP address of the computer"}
            },
            "description": "Internet connection status"
}
IP_TITLE = "internet_connection"
def _ip():
    ''' Gets the current IP and hostname for the system and returns a list of
    maps consistent with the linux DataGatherer.get_data call

    '''

    hostname = ""
    with open("/etc/hostname") as host:
        hostname = host.readline().strip()

    # many thanks to http://stackoverflow.com/a/1267524
    addresses = [(s.connect(('192.168.0.99', 80)), s.getsockname()[0], s.close()) for s in [socket.socket(socket.AF_INET, socket.SOCK_DGRAM)]]
    ip = addresses[0][1]

    resultset = {
        "hostname":hostname,
        "ip":ip
    }

    return [{"schema":IP_SCHEMA, "title":IP_TITLE, "value":resultset}]


def _active_window():
    ''' Finds the currently active window's title.

    Returns a the current window's title as a string
    '''

    window_title = ""
    try:
        # http://stackoverflow.com/a/4718297
        activewindow = subprocess.check_output(['xprop', '-root', '_NET_ACTIVE_WINDOW'])
        m = re.search('^_NET_ACTIVE_WINDOW.* ([\w]+)$', activewindow)
        if m == None:
            return []

        id_ = m.group(1)
        id_w = subprocess.check_output(['xprop', '-id', id_, 'WM_NAME'])
        match = re.match("WM_NAME\(\w+\) = (?P<name>.+)$", id_w)
        if match == None:
            return []
        window_title = match.group("name")

    except subprocess.CalledProcessError:
        # we probably don't have the call
        pass

    return window_title




class DataGatherer(object):
    def __init__(self):
        self.log_activewindow = True
        self.log_keypresses = True

    def keypresses(self):
        #Gets the number of keypresses since last call of this function
        return -1

    def windowtext(self):
        #Gets the titlebar text of currently active window
        return _active_window()

    def start(self):
        #Starts any background processes needed to get it working
        pass

    def stop(self):
        #stops logging background processes
        pass

    def get_streams(self):
        return {
            IP_TITLE:IP_SCHEMA,
            SYSINFO_TITLE:SYSINFO_SCHEMA
        }

    def get_data(self):
        ''' This function returns a list of maps of the following kind:

        {
            schema: // JsonSchema for this datapoint
            value:  // value for submitting this point
            title:  // The name that should be used for this stream
        }

        The idea is that we can gather whatever we want in the future and it'll
        all be logged here, no matter what it is.

        '''

        ip = _ip()
        sysinfo = _uname()

        all_items = ip + sysinfo

        return all_items

if __name__ == "__main__":
    dg = DataGatherer()
    print dg.windowtext()
    print dg.get_data()
