package com.connectordb.dataconnect;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;

import android.util.Log;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.GooglePlayServicesUtil;

public class SetupReceiver extends BroadcastReceiver {

    private static final String TAG = "SetupReceiver";

    public SetupReceiver() {
    }

    @Override
    public void onReceive(Context context, Intent intent) {
        // TODO: This method is called when the BroadcastReceiver is receiving
        // an Intent broadcast.
        /*On boot, start the necessary services:
        Intent locationIntent = new Intent(context, LocationService.class);
        context.startService(locationIntent);
        */

        //throw new UnsupportedOperationException("Not yet implemented");
        if (intent.getAction().equals("android.intent.action.BOOT_COMPLETED")) {
            Log.v(TAG, "RECEIVED BOOT");

            //Start all of the necessary services
            context.startService(new Intent(context, LocationService.class));
            context.startService(new Intent(context, MonitorService.class));
            context.startService(new Intent(context, FitService.class));
            context.startService(new Intent(context, SyncService.class));
        } else if (intent.getAction().equals("android.net.conn.CONNECTIVITY_CHANGE")) {
            Log.v(TAG, "CONNECTIVITY CHANGE");
            context.startService(new Intent(context, SyncService.class));
        } else {
            Log.v(TAG, "Unrecognized event: " + intent.getAction());
        }
    }

    /**
     * Check the device to make sure it has the Google Play Services APK. If
     * it doesn't, display a dialog that allows users to download the APK from
     * the Google Play Store or enable it in the device's system settings.
     */
    /*
    private final static int PLAY_SERVICES_RESOLUTION_REQUEST = 9000;
    private boolean checkPlayServices() {
        int resultCode = GooglePlayServicesUtil.isGooglePlayServicesAvailable(this);
        if (resultCode != ConnectionResult.SUCCESS) {
            if (GooglePlayServicesUtil.isUserRecoverableError(resultCode)) {
                GooglePlayServicesUtil.getErrorDialog(resultCode, this,
                        PLAY_SERVICES_RESOLUTION_REQUEST).show();
            } else {
                Log.i(TAG, "This device is not supported.");
                finish();
            }
            return false;
        }
        return true;
    }*/

    /*Should probably register for GCM here...*/
}
