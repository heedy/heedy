package com.connectordb.dataconnect;

import android.util.Log;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.location.LocationRequest;

public class GPSLogger {
    private static final String TAG = "GPSLogger";

    private LocationRequest locationRequest;
    private GoogleApiClient googleApiClient;

    public GPSLogger() {
        Log.d(TAG,"Connecting to Google Play services");

    }
}
