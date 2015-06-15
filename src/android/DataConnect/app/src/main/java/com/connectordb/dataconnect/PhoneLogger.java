package com.connectordb.dataconnect;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.os.BatteryManager;
import android.util.Log;


public class PhoneLogger {
    private static final String TAG = "PhoneLogger";

    BroadcastReceiver phoneReceiver = new BroadcastReceiver() {
        Boolean hadBatteryMessage = false;
        Boolean currentStatus = false;
        @Override
        public void onReceive(Context context, Intent intent) {
            long timestamp = System.currentTimeMillis();
            if(intent.getAction().equals(Intent.ACTION_SCREEN_ON)){
                Log.i(TAG,"screen_on: true");
            } else if(intent.getAction().equals(Intent.ACTION_SCREEN_OFF)){
                Log.i(TAG, "screen_on: false");
            } else if (intent.getAction().equals(Intent.ACTION_BATTERY_CHANGED)) {
                int plugged = intent.getIntExtra(BatteryManager.EXTRA_PLUGGED, -1);
                switch (plugged) {
                    case BatteryManager.BATTERY_PLUGGED_USB:
                    case BatteryManager.BATTERY_PLUGGED_AC:
                        if (hadBatteryMessage && !currentStatus || !hadBatteryMessage) {
                            Log.i(TAG, "plugged_in: true");
                            hadBatteryMessage = true;
                            currentStatus = true;
                        }
                        break;
                    case 0:
                        if (hadBatteryMessage && currentStatus || !hadBatteryMessage) {
                            Log.i(TAG, "plugged_in: false");
                            hadBatteryMessage = true;
                            currentStatus = false;
                        }
                        break;
                }
            }
        }
    };



    private Context mycontext;

    public PhoneLogger(Context c, int logtime) {
        mycontext = c;
        Log.d(TAG, "Registering to monitor phone metadata");
        setLogTime(logtime);
    }

    public void setLogTime(int value) {
        if (value == -1) {
            Log.i(TAG, "Disabling phone monitoring");
            mycontext.unregisterReceiver(phoneReceiver);
        } else {
            Log.i(TAG,"Enabling phone monitoring");
            IntentFilter monitorFilter = new IntentFilter();
            monitorFilter.addAction(Intent.ACTION_SCREEN_ON);
            monitorFilter.addAction(Intent.ACTION_SCREEN_OFF);
            monitorFilter.addAction(Intent.ACTION_BATTERY_CHANGED);
            mycontext.registerReceiver(phoneReceiver, monitorFilter);
        }
    }

    public void close() {
        Log.d(TAG,"Shutting down PhoneLogger");
        mycontext.unregisterReceiver(phoneReceiver);
    }
}