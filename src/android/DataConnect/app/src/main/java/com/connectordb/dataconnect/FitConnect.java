package com.connectordb.dataconnect;

import android.app.Activity;
import android.content.Context;
import android.content.IntentSender;
import android.os.AsyncTask;
import android.os.Bundle;
import android.util.Log;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.GooglePlayServicesUtil;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.common.api.PendingResult;
import com.google.android.gms.common.api.ResultCallback;
import com.google.android.gms.common.api.Status;
import com.google.android.gms.fitness.Fitness;
import com.google.android.gms.fitness.FitnessStatusCodes;
import com.google.android.gms.fitness.data.DataType;

/**
 * Created by Daniel on 4/8/2015.
 * Connects to the google fit API, and returns associated data
 */
public class FitConnect implements GoogleApiClient.ConnectionCallbacks,GoogleApiClient.OnConnectionFailedListener, ResultCallback<Status> {
    private static final String TAG = "FitClient";
    private static final int REQUEST_OAUTH = 1;

    /**
     * Track whether an authorization activity is stacking over the current activity, i.e. when
     * a known auth error is being resolved, such as showing the account chooser or presenting a
     * consent dialog. This avoids common duplications as might happen on screen rotations, etc.
     */
    private static final String AUTH_PENDING = "auth_state_pending";
    private boolean authInProgress = false;

    private GoogleApiClient mClient = null;
    private Context cont;
    private Activity act;
    FitConnect(Context c,Activity a) {
        cont = c;
        act = a;
        mClient = new GoogleApiClient.Builder(c).addConnectionCallbacks(this)
                .addOnConnectionFailedListener(this)
                .addApi(Fitness.SENSORS_API)
                .addApi(Fitness.RECORDING_API)
                .build();
        mClient.connect();
    }


    public void onConnected(Bundle connectionHint) {
        Log.i(TAG, "Connected.");
        Fitness.RecordingApi.subscribe(mClient, DataType.TYPE_ACTIVITY_SAMPLE)
                .setResultCallback(this);
        Fitness.RecordingApi.subscribe(mClient, DataType.TYPE_STEP_COUNT_DELTA)
                .setResultCallback(this);
        Fitness.RecordingApi.subscribe(mClient, DataType.TYPE_HEART_RATE_BPM)
                .setResultCallback(this);

    }

    //Subscribing to data from fitness API
    @Override
    public void onResult(Status status) {
        if (status.isSuccess()) {
            if (status.getStatusCode()
                    == FitnessStatusCodes.SUCCESS_ALREADY_SUBSCRIBED) {
                Log.i(TAG, "Existing subscription for activity detected.");
            } else {
                Log.i(TAG, "Successfully subscribed!");
            }
        } else {
            Log.i(TAG, "There was a problem subscribing.");
        }
    }

    @Override
    public void onConnectionSuspended(int cause) {
        Log.w(TAG, "Google play services connection suspended");
        // The connection has been interrupted.
        // Disable any UI components that depend on Google APIs
        // until onConnected() is called.
    }

    @Override
    public void onConnectionFailed(ConnectionResult result) {
        Log.w(TAG, "Google play services connection failed. Cause: " + result.toString());
        // This callback is important for handling errors that
        // may occur while attempting to connect with Google.
        // The failure has a resolution. Resolve it.
        // Called typically when the app is not yet authorized, and an
        // authorization dialog is displayed to the user.
        if (!authInProgress) {
            try {
                if (act !=null) {
                Log.i(TAG, "Attempting to resolve failed connection");
                authInProgress = true;
                    result.startResolutionForResult(act,
                            REQUEST_OAUTH);
                } else {
                    Log.e(TAG,"Can't resolve: must have activity");
                }
            } catch (IntentSender.SendIntentException e) {
                Log.e(TAG,
                        "Exception while starting resolution activity", e);
            }
        }
    }
}