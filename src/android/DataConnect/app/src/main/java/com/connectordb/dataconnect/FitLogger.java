package com.connectordb.dataconnect;


import android.content.Context;
import android.os.Bundle;
import android.os.Handler;
import android.util.Log;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.common.api.ResultCallback;
import com.google.android.gms.common.api.Status;
import com.google.android.gms.fitness.Fitness;
import com.google.android.gms.fitness.FitnessStatusCodes;
import com.google.android.gms.fitness.data.DataPoint;
import com.google.android.gms.fitness.data.DataType;
import com.google.android.gms.fitness.data.Field;
import com.google.android.gms.fitness.request.DataReadRequest;
import com.google.android.gms.fitness.result.DataReadResult;

import java.util.Calendar;
import java.util.Date;
import java.util.concurrent.TimeUnit;

public class FitLogger implements GoogleApiClient.ConnectionCallbacks,GoogleApiClient.OnConnectionFailedListener, ResultCallback<Status> {
    private static final String TAG = "FitLogger";
    final Handler handler = new Handler();

    private Context mycontext;
    public GoogleApiClient googleApiClient;
    public int logtime;

    public FitLogger(Context c, int logtime_) {
        mycontext = c;
        logtime = logtime_;

        googleApiClient = new GoogleApiClient.Builder(c)
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
        //LocationServices.FusedLocationApi.requestLocationUpdates(googleApiClient, locationRequest, this);
        setLogTime(logtime);
    }

    @Override
    public void onConnectionSuspended(int cause) {
        Log.w(TAG, "Google play services connection suspended");
    }
    @Override
    public void onConnectionFailed(ConnectionResult result) {
        Log.w(TAG, "Google play services connection failed. Cause: " + result.toString());

        //Call again in 5 minutes, since the user might have accepted the fit permissions dialog
        //TODO: add time delay
        Log.i(TAG,"Reconnecting in 10s");
        googleApiClient.disconnect();

        handler.postDelayed(new Runnable() {
            @Override
            public void run() {
                Log.i(TAG,"Reconnecting...");
                googleApiClient = new GoogleApiClient.Builder(mycontext)
                        .addConnectionCallbacks(FitLogger.this)
                        .addOnConnectionFailedListener(FitLogger.this)
                        .addApi(Fitness.RECORDING_API)
                        .addApi(Fitness.HISTORY_API)
                        .addScope(Fitness.SCOPE_ACTIVITY_READ)
                        .addScope(Fitness.SCOPE_BODY_READ)
                        .addScope(Fitness.SCOPE_LOCATION_READ)
                        .build();
                googleApiClient.connect();
            }
        }, 10000);

    }

    //value is milliseconds between updates - 0 is "whenever", and -1 is NONE
    public void setLogTime(int value) {
        logtime=value;
        if (value == -1) {
            if (googleApiClient.isConnected()) {
                Log.i(TAG, "Disabling google fit logger.");
                Fitness.RecordingApi.unsubscribe(googleApiClient, DataType.TYPE_ACTIVITY_SAMPLE)
                        .setResultCallback(this);
                Fitness.RecordingApi.unsubscribe(googleApiClient, DataType.TYPE_STEP_COUNT_DELTA)
                        .setResultCallback(this);
                Fitness.RecordingApi.unsubscribe(googleApiClient, DataType.TYPE_HEART_RATE_BPM)
                        .setResultCallback(this);
            }
        } else {
            Log.i(TAG, "Enabling google fit logger.");
            Fitness.RecordingApi.subscribe(googleApiClient, DataType.TYPE_ACTIVITY_SAMPLE)
                    .setResultCallback(this);
            Fitness.RecordingApi.subscribe(googleApiClient, DataType.TYPE_STEP_COUNT_DELTA)
                    .setResultCallback(this);
            Fitness.RecordingApi.subscribe(googleApiClient, DataType.TYPE_HEART_RATE_BPM)
                    .setResultCallback(this);
        }
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
            if (status.getStatusMessage()!=null) {
                Log.e(TAG, status.getStatusMessage());
            } else {
                Log.e(TAG, "subscribe to fit api failed");
            }
        }
    }



    //This stuff here gets the data (it is run in the background)
    public void getData() {
        if (!googleApiClient.isConnected()) {
            Log.w(TAG,"Can't read data: API client is not connected!");
        }
        //First get the time range for the queries on data
        Calendar cal = Calendar.getInstance();
        Date now = new Date();
        cal.setTime(now);
        long startTime = 0;
        long endTime = cal.getTimeInMillis();

        DataReadRequest readRequest = new DataReadRequest.Builder()
                .read(DataType.TYPE_ACTIVITY_SAMPLE)
                .read(DataType.TYPE_STEP_COUNT_DELTA)
                .read(DataType.TYPE_HEART_RATE_BPM)
                .setTimeRange(startTime, endTime, TimeUnit.MILLISECONDS)
                .build();

        DataReadResult dataReadResult =
                Fitness.HistoryApi.readData(googleApiClient, readRequest).await(5, TimeUnit.MINUTES);

        for (DataPoint dp : dataReadResult.getDataSet(DataType.TYPE_STEP_COUNT_DELTA).getDataPoints()) {
            //I didn't look too hard, since fuck spending more than 20 seconds to figure out how to read a damn datapoint,
            //so I did it the only way I could figure out: brute force. TL;DR: There is probably a better way of reading datapoints...
            String data = "";
            for(Field field : dp.getDataType().getFields()) {
                if (field.getName().equals("steps")) {
                    data += dp.getValue(field);
                }
            }
            endTime = dp.getEndTime(TimeUnit.MILLISECONDS);

            Log.i(TAG,"steps: "+data);
        }
        for (DataPoint dp : dataReadResult.getDataSet(DataType.TYPE_ACTIVITY_SAMPLE).getDataPoints()) {
            String data = "";
            for(Field field : dp.getDataType().getFields()) {
                if (field.getName().equals("activity")) {
                    data += dp.getValue(field).asActivity();
                }
            }
            endTime = dp.getEndTime(TimeUnit.MILLISECONDS);

            Log.i(TAG,"activity: "+data);
        }
        for (DataPoint dp : dataReadResult.getDataSet(DataType.TYPE_HEART_RATE_BPM).getDataPoints()) {
            String data = "";
            for(Field field : dp.getDataType().getFields()) {
                if (field.getName().equals("bpm")) {
                    data += dp.getValue(field);
                }
            }
            endTime = dp.getEndTime(TimeUnit.MILLISECONDS);

            Log.i(TAG,"heart_rate: "+data);
        }
    }



    public void close() {
        Log.d(TAG,"Closing Fit Logger");
        if (googleApiClient.isConnected()) {
            googleApiClient.disconnect();
        }
    }
}
