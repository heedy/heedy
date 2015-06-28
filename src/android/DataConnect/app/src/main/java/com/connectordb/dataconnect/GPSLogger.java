package com.connectordb.dataconnect;

import android.content.Context;
import android.location.Location;
import android.os.Bundle;
import android.util.Log;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.location.LocationListener;
import com.google.android.gms.location.LocationRequest;
import com.google.android.gms.location.LocationServices;

import com.connectordb.connector.Logger;

public class GPSLogger implements LocationListener, GoogleApiClient.ConnectionCallbacks,GoogleApiClient.OnConnectionFailedListener {
    private static final String TAG = "GPSLogger";

    private Context context;

    private LocationRequest locationRequest;
    private GoogleApiClient googleApiClient;

    public GPSLogger(Context c,int logtime) {
        context = c;

        Logger l = Logger.get(c);
        l.ensureStream("location", "{\"type\":\"object\",\"properties\":{\"latitude\":{\"type\":\"number\"},\"longitude\": {\"type\": \"number\"},\"altitude\": {\"type\": \"number\"},\"accuracy\": {\"type\": \"number\"},\"speed\": {\"type\": \"number\"},\"bearing\": {\"type\": \"number\"}},\"required\": [\"latitude\",\"longitude\"]}");


        Log.d(TAG, "Connecting to Google Play services");

        googleApiClient = new GoogleApiClient.Builder(c)
                .addConnectionCallbacks(this)
                .addOnConnectionFailedListener(this)
                .addApi(LocationServices.API)
                .build();

        setLogTime(logtime);
        googleApiClient.connect();
    }
    //value is milliseconds between updates - 0 is "whenever", and -1 is NONE
    public void setLogTime(int value) {
        if (value == -1) {
            Log.i(TAG,"Disabling GPS updates");
            LocationServices.FusedLocationApi.removeLocationUpdates(googleApiClient,this);
        } else if (value==0) {
            Log.i(TAG,"Setting Battery Saver mode");
            locationRequest = new LocationRequest();
            locationRequest.setFastestInterval(1000);
            locationRequest.setPriority(LocationRequest.PRIORITY_NO_POWER);
        } else {
            Log.i(TAG, "Setting location ms: " + Integer.toString(value));
            locationRequest = new LocationRequest();
            locationRequest.setInterval(value);
            locationRequest.setFastestInterval(1000);
            locationRequest.setPriority(LocationRequest.PRIORITY_HIGH_ACCURACY);
        }
    }

    @Override
    public void onConnected(Bundle connectionHint) {
        Log.d(TAG,"Google play services connected.");
        LocationServices.FusedLocationApi.requestLocationUpdates(googleApiClient, locationRequest, this);
    }

    @Override
    public void onConnectionSuspended(int cause) {
        Log.w(TAG, "Google play services connection suspended");
    }
    @Override
    public void onConnectionFailed(ConnectionResult result) {
        Log.w(TAG, "Google play services connection failed");
    }

    @Override
    public void onLocationChanged(Location location) {
        //Called when the location changes
        String data = "{\"latitude\": "+Double.toString(location.getLatitude())+
                ", \"longitude\": "+Double.toString(location.getLongitude());
        if (location.hasAltitude()) {
            data += ", \"altitude\": " + Double.toString(location.getAltitude());
        }
        if (location.hasAccuracy()) {
            data+= ", \"accuracy\": "+Double.toString(location.getAccuracy());
        }
        if (location.hasSpeed()) {
            data+= ", \"speed\": " + Double.toString(location.getSpeed());
        }
        if (location.hasBearing()) {
            data+= ", \"bearing\": " + Double.toString(location.getBearing());
        }

        Logger.get(context).Insert("location",location.getTime(),data+"}");

    }

    public void close() {
        Log.d(TAG,"Closing GPS Logger");
        if (googleApiClient.isConnected()) {
            LocationServices.FusedLocationApi.removeLocationUpdates(googleApiClient,this);
            googleApiClient.disconnect();
        }

    }
}
