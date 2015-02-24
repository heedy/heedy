package com.connectordb.dataconnect;

import android.app.Service;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.os.BatteryManager;
import android.os.IBinder;
import android.preference.PreferenceManager;
import android.util.Log;

public class MonitorService extends Service{

    private static final String TAG = "MonitorService";

    BroadcastReceiver mMonitorReceiver = new BroadcastReceiver() {
        //The plugged_in part is repeated many times in a row - so it makes no sense to write it many times.
        Boolean hadBatteryMessage = false;
        Boolean currentStatus = false;
        @Override
        public void onReceive(Context context, Intent intent) {
            long timestamp = System.currentTimeMillis();

            if(intent.getAction().equals(Intent.ACTION_SCREEN_ON)){
                DataCache.get(MonitorService.this).Insert("screen_on", timestamp, "true");
            } else if(intent.getAction().equals(Intent.ACTION_SCREEN_OFF)){
                DataCache.get(MonitorService.this).Insert("screen_on", timestamp, "false");

            } else if (intent.getAction().equals(Intent.ACTION_BATTERY_CHANGED)) {
                int plugged = intent.getIntExtra(BatteryManager.EXTRA_PLUGGED, -1);
                switch (plugged) {
                    case BatteryManager.BATTERY_PLUGGED_USB:
                    case BatteryManager.BATTERY_PLUGGED_AC:
                        if (hadBatteryMessage && !currentStatus || !hadBatteryMessage) {
                            DataCache.get(MonitorService.this).Insert("plugged_in", timestamp, "true");
                            hadBatteryMessage = true;
                            currentStatus = true;
                        }
                        break;
                    case 0:
                        if (hadBatteryMessage && currentStatus || !hadBatteryMessage) {
                            DataCache.get(MonitorService.this).Insert("plugged_in", timestamp, "false");
                            hadBatteryMessage = true;
                            currentStatus = false;
                        }
                        break;
                }

            }

        }
    };

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Boolean value = PreferenceManager.getDefaultSharedPreferences(this).getBoolean("monitor_enabled", true);
        try {
            Boolean newvalue = intent.getBooleanExtra("enabled", true);
            if (newvalue!=value) {
                Log.v(TAG, "Updating service enabled: " + newvalue.toString());
                PreferenceManager.getDefaultSharedPreferences(this).edit().putBoolean("monitor_enabled", newvalue).commit();
                value = newvalue;
            }
        } catch (NullPointerException ex) {
            Log.v(TAG,"No data detected. Using existing settings.");
        }

        //There we go, shutdown if it is started wrong
        if (!value) {
            stopSelf();
            return START_NOT_STICKY;
        }
        return START_STICKY;
    }

    @Override
    public void onCreate() {
        Log.i(TAG, "Registering to monitor screen/battery");
        IntentFilter monitorFilter = new IntentFilter();
        monitorFilter.addAction(Intent.ACTION_SCREEN_ON);
        monitorFilter.addAction(Intent.ACTION_SCREEN_OFF);
        monitorFilter.addAction(Intent.ACTION_BATTERY_CHANGED);
        registerReceiver(mMonitorReceiver, monitorFilter);

    }

    @Override
    public void onDestroy() {
        Log.i(TAG,"Destroy service");
        unregisterReceiver(mMonitorReceiver);
    }
}
