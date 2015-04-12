package com.connectordb.dataconnect;

import android.app.Service;
import android.content.Intent;
import android.location.Location;
import android.os.Bundle;
import android.os.IBinder;

import android.preference.PreferenceManager;
import android.util.Log;

//import com.google.android.gms.common.GooglePlayServicesClient;
import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.common.api.GoogleApiClient.ConnectionCallbacks;
import com.google.android.gms.location.LocationListener;
import com.google.android.gms.location.LocationRequest;
import com.google.android.gms.location.LocationServices;

public class LocationService extends Service implements LocationListener, ConnectionCallbacks, GoogleApiClient.OnConnectionFailedListener {

    private static final String TAG = "LocationService";

    private LocationRequest mLocationRequest;
    private GoogleApiClient mGoogleApiClient;

    @Override
    public void onLocationChanged(Location location)
    {
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
        DataCache.get(this).Insert("gps",location.getTime(),data+"}");
    }


    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    private void makeRequest(int value) {
        if (value == -1) {
            stopSelf();
        } else if (value == 0) {
            Log.i(TAG, "Setting Battery Saver Mode");
            mLocationRequest = new LocationRequest();
            mLocationRequest.setFastestInterval(100);
            mLocationRequest.setPriority(LocationRequest.PRIORITY_NO_POWER);
        } else {
            Log.i(TAG, "Setting location ms: " + Integer.toString(value));
            mLocationRequest = new LocationRequest();
            mLocationRequest.setInterval(value);
            mLocationRequest.setFastestInterval(5000);
            mLocationRequest.setPriority(LocationRequest.PRIORITY_HIGH_ACCURACY);
        }
        if (mGoogleApiClient.isConnected()) {
            LocationServices.FusedLocationApi.requestLocationUpdates(
                    mGoogleApiClient, mLocationRequest, this);
        } else {
            Log.v(TAG, "Not connected to google play - can't request GPS updates.");
        }
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        int value = PreferenceManager.getDefaultSharedPreferences(this).getInt("location_update_frequency",0);
        int newvalue;
        try {
            newvalue = intent.getIntExtra("location_update_frequency", -2);
        } catch (NullPointerException ex) {
            newvalue = -2;
        }

        if (newvalue >= -1 && newvalue!=value) {
            Log.v(TAG,"Updating GPS settings: "+Integer.toString(value)+"->"+Integer.toString(newvalue));
            //Update the value in the settings
            PreferenceManager.getDefaultSharedPreferences(this).edit().putInt("location_update_frequency",newvalue).commit();
            value = newvalue;

        }
        makeRequest(value);


        return START_STICKY;
    }

    @Override
    public void onCreate() {
        Log.i(TAG,"Connecting to google play services");

        mLocationRequest = new LocationRequest();
        mLocationRequest.setFastestInterval(100);
        mLocationRequest.setPriority(LocationRequest.PRIORITY_NO_POWER);

        mGoogleApiClient = new GoogleApiClient.Builder(this)
                .addConnectionCallbacks(this)
                .addOnConnectionFailedListener(this)
                .addApi(LocationServices.API)
                .build();
        mGoogleApiClient.connect();


    }
    public void onConnected(Bundle connectionHint) {
        Log.i(TAG,"Connected. Requesting GPS updates.");
        LocationServices.FusedLocationApi.requestLocationUpdates(
                mGoogleApiClient, mLocationRequest, this);
    }
    @Override
    public void onConnectionSuspended(int cause) {
        Log.w(TAG,"Google play services connection suspended");
        // The connection has been interrupted.
        // Disable any UI components that depend on Google APIs
        // until onConnected() is called.
    }
    @Override
    public void onConnectionFailed(ConnectionResult result) {
        Log.w(TAG,"Google play services connection failed");
        // This callback is important for handling errors that
        // may occur while attempting to connect with Google.
        //
        // More about this in the next section.

    }
    @Override
    public void onDestroy() {
        if (mGoogleApiClient.isConnected()) {
            LocationServices.FusedLocationApi.removeLocationUpdates(
                    mGoogleApiClient, this);
            mGoogleApiClient.disconnect();
        }
        Log.i(TAG,"Destroy service");
    }
}
