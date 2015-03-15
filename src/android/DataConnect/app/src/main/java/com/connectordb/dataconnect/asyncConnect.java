package com.connectordb.dataconnect;

import android.os.AsyncTask;
import android.util.Log;

/**
 * Created by Daniel on 3/15/2015.
 */
public class asyncConnect extends AsyncTask<String, Integer, Long> {
    private static final String TAG = "asyncConnect";
    protected Long doInBackground(String... data) { //This is shit - don't have time to figure out how it works
        ConnectorDB cdb = new ConnectorDB(data[0],data[1],data[2]);
        String devicename = android.os.Build.MODEL.replaceAll(" ","");
        cdb.makedevice(devicename);
        cdb.makestream(devicename,"gps","f[2]/gps");
        Log.d(TAG, "USERS DONE");
        return Long.valueOf(0);
            }

    protected void onProgressUpdate(Integer... progress) {

            }

    protected void onPostExecute(Long result) {
            }
}