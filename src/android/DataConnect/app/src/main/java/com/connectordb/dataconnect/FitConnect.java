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
import com.google.android.gms.fitness.data.DataPoint;
import com.google.android.gms.fitness.data.DataSet;
import com.google.android.gms.fitness.data.DataType;
import com.google.android.gms.fitness.data.Field;
import com.google.android.gms.fitness.request.DataReadRequest;
import com.google.android.gms.fitness.result.DataReadResult;

import java.text.SimpleDateFormat;
import java.util.Calendar;
import java.util.Date;
import java.util.concurrent.TimeUnit;

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
    private Boolean hadSubscribe = false;
    private Boolean isSubscribe = true;

    FitConnect(Context c,Activity a,Boolean isSubscribed) {
        cont = c;
        act = a;
        isSubscribe = isSubscribed;
        mClient = new GoogleApiClient.Builder(cont).addConnectionCallbacks(this)
                .addOnConnectionFailedListener(this)
                .addApi(Fitness.RECORDING_API)
                .addApi(Fitness.HISTORY_API)
                .addScope(Fitness.SCOPE_ACTIVITY_READ)
                .addScope(Fitness.SCOPE_BODY_READ)
                .addScope(Fitness.SCOPE_LOCATION_READ)
                .build();
        mClient.connect();
    }

    public void reconnect() {

        if (!mClient.isConnected() && !mClient.isConnecting()) {
            Log.i(TAG,"Reconnecting");
            mClient.connect();
        }
    }

    public void subscribe() {
        hadSubscribe=true;
        isSubscribe = true;

        if (mClient!=null) {
            if (mClient.isConnected()) {
                Log.i(TAG,"subscribing");
                Fitness.RecordingApi.subscribe(mClient, DataType.TYPE_ACTIVITY_SAMPLE)
                        .setResultCallback(this);
                Fitness.RecordingApi.subscribe(mClient, DataType.TYPE_STEP_COUNT_DELTA)
                        .setResultCallback(this);
                Fitness.RecordingApi.subscribe(mClient, DataType.TYPE_HEART_RATE_BPM)
                        .setResultCallback(this);
            }
        }
    }

    public void unsubscribe() {
        hadSubscribe=true;
        isSubscribe = false;
        if (mClient!=null) {
            if (mClient.isConnected()) {
                Log.i(TAG,"unsubscribing");
                Fitness.RecordingApi.unsubscribe(mClient, DataType.TYPE_ACTIVITY_SAMPLE)
                        .setResultCallback(this);
                Fitness.RecordingApi.unsubscribe(mClient, DataType.TYPE_STEP_COUNT_DELTA)
                        .setResultCallback(this);
                Fitness.RecordingApi.unsubscribe(mClient, DataType.TYPE_HEART_RATE_BPM)
                        .setResultCallback(this);
            }
        }
    }

    //Writes the data to the data cache - and returns the end time of the data
    public void getdata() {
        writeActivitySample();
        writeStepCount();
        writeHeartRate();

    }

    public void writeActivitySample() {
        Calendar cal = Calendar.getInstance();
        Date now = new Date();
        cal.setTime(now);
        String longtext = DataCache.get(cont).GetKey("fit_starttime_activity");
        if (longtext.length()==0) {
            longtext = "2000";
        }
        long startTime = Long.parseLong(longtext);
        long endTime = cal.getTimeInMillis();
        Log.v(TAG,"Start time: "+startTime);
        DataReadRequest readRequest = new DataReadRequest.Builder()
                .read(DataType.TYPE_ACTIVITY_SAMPLE)
                .setTimeRange(startTime, endTime, TimeUnit.MILLISECONDS)
                .build();
        // Invoke the History API to fetch the data with the query and await the result of
        // the read request.
        DataReadResult dataReadResult =
                Fitness.HistoryApi.readData(mClient, readRequest).await(1, TimeUnit.MINUTES);
        endTime = 0;
        for (DataPoint dp : dataReadResult.getDataSet(DataType.TYPE_ACTIVITY_SAMPLE).getDataPoints()) {
            String data = "{";
            for(Field field : dp.getDataType().getFields()) {
                if (field.getName().equals("activity")) {
                    data += "\"" + field.getName() + "\": \"" + dp.getValue(field).asActivity() + "\",";
                } else{
                    data += "\"" + field.getName() + "\": " + dp.getValue(field) + ",";
                }
            }
            DataCache.get(cont).Insert("activity_name", dp.getEndTime(TimeUnit.MILLISECONDS), data.substring(0, data.length() - 1) + "}");

            if (dp.getEndTime(TimeUnit.MILLISECONDS)>endTime) {
                endTime = dp.getEndTime(TimeUnit.MILLISECONDS);
            }
        }

        Log.v(TAG,"Endtime:"+endTime);
        if (endTime >= startTime) {
            DataCache.get(cont).SetKey("fit_starttime_activity",Long.toString(endTime+1));
        }
    }
    public void writeStepCount() {
        Calendar cal = Calendar.getInstance();
        Date now = new Date();
        cal.setTime(now);
        String longtext = DataCache.get(cont).GetKey("fit_starttime_steps");
        if (longtext.length()==0) {
            longtext = "2000";
        }
        long startTime = Long.parseLong(longtext);
        long endTime = cal.getTimeInMillis();
        DataReadRequest readRequest = new DataReadRequest.Builder()
                .read(DataType.TYPE_STEP_COUNT_DELTA)
                .setTimeRange(startTime, endTime, TimeUnit.MILLISECONDS)
                .build();
        // Invoke the History API to fetch the data with the query and await the result of
        // the read request.
        DataReadResult dataReadResult =
                Fitness.HistoryApi.readData(mClient, readRequest).await(5, TimeUnit.MINUTES);
        endTime = 0;
        Log.v(TAG,"Start time: "+startTime);
        for (DataPoint dp : dataReadResult.getDataSet(DataType.TYPE_STEP_COUNT_DELTA).getDataPoints()) {
            String data = "{";
            for(Field field : dp.getDataType().getFields()) {
                data += "\""+field.getName()+"\": "+dp.getValue(field)+",";
            }
            DataCache.get(cont).Insert("stepcount", dp.getEndTime(TimeUnit.MILLISECONDS), data.substring(0, data.length() - 1) + "}");

            if (dp.getEndTime(TimeUnit.MILLISECONDS)>endTime) {
                endTime = dp.getEndTime(TimeUnit.MILLISECONDS);
            }


        }
        if (endTime >= startTime) {
            DataCache.get(cont).SetKey("fit_starttime_steps", Long.toString(endTime+1));
        }
    }
    public void writeHeartRate() {
        Calendar cal = Calendar.getInstance();
        Date now = new Date();
        cal.setTime(now);
        String longtext = DataCache.get(cont).GetKey("fit_starttime_heartrate");
        if (longtext.length()==0) {
            longtext = "2000";
        }
        long startTime = Long.parseLong(longtext);
        Log.v(TAG,"Start time: "+startTime);
        long endTime = cal.getTimeInMillis();
        DataReadRequest readRequest = new DataReadRequest.Builder()
                .read(DataType.TYPE_HEART_RATE_BPM)
                .setTimeRange(startTime, endTime, TimeUnit.MILLISECONDS)
                .build();
        // Invoke the History API to fetch the data with the query and await the result of
        // the read request.
        DataReadResult dataReadResult =
                Fitness.HistoryApi.readData(mClient, readRequest).await(5, TimeUnit.MINUTES);
        endTime = 0;
        for (DataPoint dp : dataReadResult.getDataSet(DataType.TYPE_HEART_RATE_BPM).getDataPoints()) {
            String data = "{";
            for(Field field : dp.getDataType().getFields()) {
                data += "\""+field.getName()+"\": "+dp.getValue(field)+",";
            }
            DataCache.get(cont).Insert("heartrate", dp.getEndTime(TimeUnit.MILLISECONDS), data.substring(0, data.length() - 1) + "}");

            if (dp.getEndTime(TimeUnit.MILLISECONDS)>endTime) {
                endTime = dp.getEndTime(TimeUnit.MILLISECONDS);
            }
        }
        if (endTime > startTime) {
            DataCache.get(cont).SetKey("fit_starttime_heartrate",Long.toString(endTime+1));
        }
    }


    public void onConnected(Bundle connectionHint) {
        Log.i(TAG, "Connected.");
        if (hadSubscribe) {
            if (isSubscribe) {
                subscribe();
            } else {
                unsubscribe();
            }
        }

    }

    //Subscribing to data from fitness API
    @Override
    public void onResult(Status status) {
        if (status.isSuccess()) {
            Log.i(TAG, "subscription success" );
        } else {
            if (status.getStatusMessage()!=null) {
                Log.e(TAG, status.getStatusMessage());
            } else {
                Log.e(TAG, "subscribe failed");
            }
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