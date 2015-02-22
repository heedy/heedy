package com.connectordb.dataconnect;

import android.app.Service;
import android.content.Intent;
import android.location.Location;
import android.os.Binder;
import android.os.Bundle;
import android.os.IBinder;
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
        Log.d(TAG, "GPS: " + location.toString() +"speed=" + location.getSpeed()+" t=" + location.getTime());
    }


    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        String stringValue = intent.getStringExtra("value");
        int value = Integer.parseInt(stringValue);
        if (value==-1) {
            stopSelf();
            return START_STICKY;
        } else if (value==0) {
            Log.i(TAG, "Setting Battery Saver Mode");
            mLocationRequest = new LocationRequest();
            mLocationRequest.setFastestInterval(100);
            mLocationRequest.setPriority(LocationRequest.PRIORITY_NO_POWER);
        } else {
            Log.i(TAG, "Setting location ms: "+stringValue);
            mLocationRequest = new LocationRequest();
            mLocationRequest.setInterval(value);
            mLocationRequest.setFastestInterval(5000);
            mLocationRequest.setPriority(LocationRequest.PRIORITY_HIGH_ACCURACY);
        }
        if (mGoogleApiClient.isConnected()) {
            LocationServices.FusedLocationApi.requestLocationUpdates(
                    mGoogleApiClient, mLocationRequest, this);
        } else {
            Log.v(TAG, "Not connected to google play - can't update request.");
        }
        // We want this service to continue running until it is explicitly
        // stopped, so return sticky.
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
        Log.i(TAG,"Requesting GPS updates.");
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
