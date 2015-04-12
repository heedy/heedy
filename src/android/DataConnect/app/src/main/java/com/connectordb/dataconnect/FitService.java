package com.connectordb.dataconnect;

import android.app.Service;
import android.content.Intent;
import android.os.IBinder;
import android.preference.PreferenceManager;
import android.util.Log;

import com.google.android.gms.common.api.GoogleApiClient;

import java.util.Timer;
import java.util.TimerTask;

public class FitService extends Service {

    private static final String TAG = "FitService";
    FitConnect fit;
    private Timer timer = new Timer();
    public FitService() {
    }

    @Override
    public IBinder onBind(Intent intent) {return null;}



    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
            int value = PreferenceManager.getDefaultSharedPreferences(this).getInt("fit_update_frequency",0);
            int newvalue;
            try {
            newvalue = intent.getIntExtra("fit_update_frequency", -2);
            } catch (NullPointerException ex) {
            newvalue = -2;
            }

            if (newvalue >= -1 && newvalue!=value) {
                Log.v(TAG,"Updating fit settings: "+Integer.toString(value)+"->"+Integer.toString(newvalue));
                //Update the value in the settings
                PreferenceManager.getDefaultSharedPreferences(this).edit().putInt("fit_update_frequency",newvalue).commit();
                value = newvalue;
                timer.cancel();
                if (newvalue >= 0) {
                    fit.subscribe();
                    if (newvalue > 0) {
                        timer = new Timer();
                        timer.scheduleAtFixedRate(new TimerTask() {
                            @Override
                            public void run() {
                                Log.v(TAG,"Fit: Get data.");
                                FitService.this.fit.getdata();
                            }
                        },1000,newvalue);
                    } else {
                        return START_NOT_STICKY;
                    }
                } else {
                    fit.unsubscribe();
                    return START_NOT_STICKY;
                }

            }
            return START_STICKY;
            }

    @Override
    public void onCreate() {
            Log.i(TAG,"Connecting to google play services");
            fit = new FitConnect(this,null,true);
     }

    @Override
    public void onDestroy() {
        fit.disconnect();
        timer.cancel();
    }
}
