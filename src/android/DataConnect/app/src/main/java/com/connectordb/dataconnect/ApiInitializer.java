package com.connectordb.dataconnect;

import android.app.Activity;
import android.content.IntentSender;
import android.os.Bundle;
import android.util.Log;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.fitness.Fitness;

public class ApiInitializer implements GoogleApiClient.ConnectionCallbacks,GoogleApiClient.OnConnectionFailedListener {
    private static final String TAG = "ApiInitializer";
    private static final int REQUEST_OAUTH = 1;

    /**
     * Track whether an authorization activity is stacking over the current activity, i.e. when
     * a known auth error is being resolved, such as showing the account chooser or presenting a
     * consent dialog. This avoids common duplications as might happen on screen rotations, etc.
     */
    private static final String AUTH_PENDING = "auth_state_pending";


    Activity activity;
    public GoogleApiClient googleApiClient;

    public ApiInitializer(Activity a) {
        activity = a;

        googleApiClient = new GoogleApiClient.Builder(activity)
                .addConnectionCallbacks(this)
                .addOnConnectionFailedListener(this)
                .addApi(Fitness.RECORDING_API)
                .addApi(Fitness.HISTORY_API)
                .addScope(Fitness.SCOPE_ACTIVITY_READ)
                .addScope(Fitness.SCOPE_BODY_READ)
                .addScope(Fitness.SCOPE_LOCATION_READ)
                .build();

        googleApiClient.connect();
    }

    @Override
    public void onConnected(Bundle connectionHint) {
        Log.d(TAG, "Google play services connected.");
    }

    @Override
    public void onConnectionSuspended(int cause) {
        Log.w(TAG, "Google play services connection suspended");
    }
    @Override
    public void onConnectionFailed(ConnectionResult result) {
        Log.w(TAG, "Google play services connection failed. Cause: " + result.toString());
        // This callback is important for handling errors that
        // may occur while attempting to connect with Google.
        // The failure has a resolution. Resolve it.
        // Called typically when the app is not yet authorized, and an
        // authorization dialog is displayed to the user.

        try {

            Log.i(TAG, "Attempting to resolve failed connection");

            result.startResolutionForResult(activity,
                    REQUEST_OAUTH);

        } catch (IntentSender.SendIntentException e) {
            Log.e(TAG,
                    "Exception while starting resolution activity", e);
        }

    }

    public void reconnect() {
        Log.v(TAG,"Reconnecting...");
        googleApiClient.reconnect();
    }
}
