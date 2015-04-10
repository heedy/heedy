package com.connectordb.dataconnect;

import android.content.Context;
import android.database.Cursor;
import android.os.AsyncTask;
import android.util.Log;
import android.widget.Toast;

import com.google.gson.JsonObject;

/**
 * Created by Daniel on 3/15/2015.
 */
public class asyncConnect extends AsyncTask<String, Integer, Long> {
    private static final String TAG = "asyncConnect";
    Context context;
    asyncConnect(Context c) {
        this.context = c;
    }
    protected Long doInBackground(String... data) { //This is shit - don't have time to figure out how it works

        FitConnect fit = new FitConnect(this.context,null,true);

        fit.getdata();

        /*
        ConnectorDB cdb = new ConnectorDB(data[0],data[1],data[2]);
        String devicename = android.os.Build.MODEL.replaceAll(" ","");
        cdb.makedevice(devicename);
        cdb.makestream(devicename,"gps","f[2]/gps");
        cdb.makestream(devicename,"plugged_in","b");
        cdb.makestream(devicename,"screen_on","b");

        DataCache c = DataCache.get(this.context);
        //Now add ALL the data
        Cursor cursor = c.getCache();
        boolean val = cursor.moveToFirst();
        double prevtime = 1.424813077181E9;
        while (val) {
            long id = cursor.getLong(0);
            double timestamp = ((double)cursor.getLong(1))/1000;


            String stream = cursor.getString(2);
            String dataval = cursor.getString(3);
            String jsonInsert = "{\"T\":"+Double.toString(timestamp)+",\"D\":"+dataval+"}";
            Log.d(TAG, jsonInsert);
            while (timestamp<=prevtime) {
                timestamp += 1e-4;
                jsonInsert = "{\"T\":"+Double.toString(timestamp)+",\"D\":"+dataval+"}";
                Log.d(TAG, "CORRECTED TIME: "+jsonInsert);
            }


            if (cdb.insert(devicename,stream,jsonInsert)) {
                //The insert was successful! Delete the value from the database
                c.Delete(id);
            } else {
                //The insert failed.

                //Check to see if the insert failed because of a duplicate timestamp
                timestamp += 1e-4;
                jsonInsert = "{\"T\":"+Double.toString(timestamp)+",\"D\":"+dataval+"}";
                Log.d(TAG, "ATTEMPT CORRECT TIME: "+jsonInsert);
                if (cdb.insert(devicename,stream,jsonInsert)) {
                    //The insert was successful! Delete the value from the database
                    c.Delete(id);
                } else {
                    //Nope - something is fukt
                    return Long.valueOf(1);
                }
            }
            prevtime=timestamp;
            val = cursor.moveToNext();
        }
        */

        Log.d(TAG, "SYNC DONE");

        return Long.valueOf(0);
            }

    protected void onProgressUpdate(Integer... progress) {

            }

    protected void onPostExecute(Long result) {
        if (result==0) {
            Toast.makeText(context, "Sync Finished", Toast.LENGTH_SHORT).show();
        } else {
            Toast.makeText(context, "Sync Failed", Toast.LENGTH_SHORT).show();
        }
            }
}