package com.connectordb.dataconnect;

import android.app.Service;
import android.content.Intent;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.IBinder;
import android.preference.PreferenceManager;
import android.util.Log;

import java.util.Timer;
import java.util.TimerTask;

public class SyncService extends Service {

    private static final String TAG = "SyncService";
    private Timer timer = new Timer();
    private Boolean iswifi = false;

    public SyncService() {
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }


    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        int wifivalue = PreferenceManager.getDefaultSharedPreferences(this).getInt("wifi_sync_update_frequency", 0);
        int mobilevalue = PreferenceManager.getDefaultSharedPreferences(this).getInt("mobile_sync_update_frequency", 0);

        try {
            int newvalue = intent.getIntExtra("wifi_sync_update_frequency", -2);
            if (newvalue!=-2) {
                PreferenceManager.getDefaultSharedPreferences(this).edit().putInt("wifi_sync_update_frequency", newvalue).commit();
                Log.v(TAG, "Updating wifi sync settings: " + Integer.toString(wifivalue) + "->" + Integer.toString(newvalue));
                wifivalue = newvalue;
            }
        } catch (NullPointerException ex) {

        }

        try {
            int newvalue = intent.getIntExtra("mobile_sync_update_frequency", -2);
            if (newvalue!=-2) {
                PreferenceManager.getDefaultSharedPreferences(this).edit().putInt("mobile_sync_update_frequency", newvalue).commit();
                Log.v(TAG, "Updating mobile sync settings: " + Integer.toString(mobilevalue) + "->" + Integer.toString(newvalue));
                mobilevalue = newvalue;
            }
        } catch (NullPointerException ex) {

        }

        ConnectivityManager connManager = (ConnectivityManager) getSystemService(this.CONNECTIVITY_SERVICE);
        NetworkInfo mWifi = connManager.getNetworkInfo(ConnectivityManager.TYPE_WIFI);

        timer.cancel();
        if (mWifi.isConnected()) {
            Log.v(TAG, "Running on wifi settings");
            if (wifivalue > 0) {
                timer = new Timer();
                timer.scheduleAtFixedRate(new TimerTask() {
                    @Override
                    public void run() {
                        Log.v(TAG, "Syncing Data");
                        ConnectorDB.get(SyncService.this).Sync();
                    }
                }, wifivalue, wifivalue);
                return START_STICKY;
            }
        } else {
            Log.v(TAG, "Running on mobile settings");
            if (mobilevalue > 0) {
                timer = new Timer();
                timer.scheduleAtFixedRate(new TimerTask() {
                    @Override
                    public void run() {
                        Log.v(TAG, "Syncing Data");
                        ConnectorDB.get(SyncService.this).Sync();
                    }
                }, mobilevalue, mobilevalue);
                return START_STICKY;
            }
        }

        return START_NOT_STICKY;
    }

    @Override
    public void onCreate() {
    }

    @Override
    public void onDestroy() {
        timer.cancel();
    }
}
