package com.connectordb.dataconnect;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.util.Log;

public class DataReceiver extends BroadcastReceiver {
    private static final String TAG = "DATAReceiver";
    public DataReceiver() {
    }

    @Override
    public void onReceive(Context context, Intent intent) {
        // TODO: This method is called when the BroadcastReceiver is receiving
        // an Intent broadcast.
        if (intent.getAction().equals("android.intent.action.BOOT_COMPLETED")) {
            Log.v(TAG, "RECEIVED BOOT");

            //Start the logger service
            context.startService(new Intent(context, LoggerService.class));

        } else if (intent.getAction().equals("android.net.conn.CONNECTIVITY_CHANGE")) {
            Log.v(TAG, "CONNECTIVITY CHANGE");
            //context.startService(new Intent(context, SyncService.class));
        } else {
            Log.w(TAG, "Unrecognized event: " + intent.getAction());
        }
    }
}
