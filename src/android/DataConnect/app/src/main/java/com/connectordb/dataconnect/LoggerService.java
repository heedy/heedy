package com.connectordb.dataconnect;

import android.app.Service;
import android.content.Intent;
import android.os.IBinder;
import android.util.Log;

public class LoggerService extends Service {
    private static final String TAG = "LoggerService";

    public LoggerService() {
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    //The list of active loggers
    public GPSLogger gpslogger;
    public PhoneLogger phonelogger;

    @Override
    public void onCreate() {
        Log.d(TAG,"Initializing loggers...");

        gpslogger = new GPSLogger(this,10000);
        phonelogger = new PhoneLogger(this,1);
    }

    @Override
    public void onDestroy() {
        Log.d(TAG,"Shutting down logger service");
        gpslogger.close();
        phonelogger.close();
    }
}
