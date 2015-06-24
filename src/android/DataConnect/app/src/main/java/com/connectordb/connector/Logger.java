package com.connectordb.connector;

import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.DatabaseUtils;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;
import android.os.AsyncTask;
import android.os.Handler;
import android.util.Log;

public class Logger extends SQLiteOpenHelper {
    public static final int DATABASE_VERSION = 2;
    public static final String TAG = "ConnectorDB:Logger";
    public static final String DATABASE_NAME = "ConnectorLog.db";

    final Handler handler = new Handler();
    Runnable syncer = new Runnable() {
        public void run() {
            new AsyncTask<Void, Void, Void>() {
                @Override
                protected Void doInBackground(Void... params) {
                    Logger.this.Sync();
                    Logger.this.StartSyncWait();
                    return null;
                }
            }.execute();
        }
    };

    //The Logger class is used as a singleton in the application
    private static Logger logger_singleton;
    public static synchronized Logger get(Context c) {
        if (logger_singleton==null) {
            if (c==null) {
                Log.e(TAG,"Context not supplied to logger!");
            }
            Log.v(TAG, "Initializing logger");
            logger_singleton = new Logger(c);
        }
        return logger_singleton;
    }

    public Logger(Context context) {
        super(context, DATABASE_NAME, null, DATABASE_VERSION);

        long syncenabled = 0;
        try {
            syncenabled = Long.parseLong(this.GetKey("syncenabled"));
        } catch(NumberFormatException nfe) {}

        if (syncenabled > 0) {
            Log.i(TAG, "Logger init sync enable");
            this.StartSyncWait();
        }

    }

    @Override
    public void onCreate(SQLiteDatabase db) {
        Log.v(TAG, "Creating new logger cache database");
        db.execSQL("CREATE TABLE streams (streamname TEXT PRIMARY KEY, schema TEXT);");
        db.execSQL("CREATE TABLE cache (streamname TEXT, timestamp REAL, data TEXT);");
        db.execSQL("CREATE TABLE kv (key TEXT PRIMARY KEY, value TEXT);");

        //Now fill in the default values in kv
        db.execSQL("INSERT INTO kv VALUES ('server','https://connectordb.com');");
        db.execSQL("INSERT INTO kv VALUES ('devicename','');");
        db.execSQL("INSERT INTO kv VALUES ('__apikey','');");
        db.execSQL("INSERT INTO kv VALUES ('syncperiod','3600');"); //Make the database sync every hour
        db.execSQL("INSERT INTO kv VALUES ('syncenabled','0');"); //Make the database sync every hour
    }

    @Override
    public void onUpgrade(SQLiteDatabase db, int oldVersion, int newVersion) {
        Log.w(TAG, "Upgrading Logger - deleting old tables...");
        db.execSQL("DROP TABLE IF EXISTS cache;");
        db.execSQL("DROP TABLE IF EXISTS streams;");
        db.execSQL("DROP TABLE IF EXISTS kv;");
        onCreate(db);
    }

    public void ensureStream(String stream,String schema) {
        Log.v(TAG, "Ensuring stream " + stream);

        SQLiteDatabase db = this.getWritableDatabase();
        ContentValues contentValues = new ContentValues();
        contentValues.put("streamname", stream);
        contentValues.put("schema", schema);
        db.insertWithOnConflict("streams", null, contentValues, SQLiteDatabase.CONFLICT_REPLACE);
    }

    //Inserts a datapoint to the stream (jsonified)
    public boolean Insert(String stream, long timestamp, String data) {
        Log.v(TAG, "[s=" + stream + " t=" + Long.toString(timestamp) + " d=" + data + "]");

        SQLiteDatabase db = this.getWritableDatabase();
        ContentValues contentValues = new ContentValues();
        contentValues.put("streamname", stream);
        contentValues.put("timestamp", ((double) timestamp) / 1000.0);
        contentValues.put("data", data);
        db.insert("cache", null, contentValues);
        return true;
    }

    //Returns the number of cached datapoints
    public int Size() {
        SQLiteDatabase db = this.getReadableDatabase();
        int numRows = (int) DatabaseUtils.queryNumEntries(db, "cache");
        Log.v(TAG, "Cache Size: " + Integer.toString(numRows));
        return numRows;
    }

    public void StartSyncWait() {


        long waittime = Long.parseLong(this.GetKey("syncperiod"));

        if (waittime > 0) {
            Log.v(TAG,"Setting next sync in "+ waittime);
            handler.postDelayed(syncer,waittime);
        }
    }

    public void DisableTimedSync() {
        Log.v(TAG, "Disabling syncer");
        handler.removeCallbacks(syncer);
        this.SetKey("syncenabled", "0");

    }

    public void EnableTimedSync(long time) {
        DisableTimedSync();
        this.SetKey("syncenabled", "1");
        this.SetKey("syncperiod",Long.toString(time));
        StartSyncWait();
    }

    public void BGSync() {
        new AsyncTask<Void, Void, Void>() {
            @Override
            protected Void doInBackground(Void... params) {
                Logger.this.Sync();
                return null;
            }
        }.execute();
    }

    //Synchronizes the database with the server
    public synchronized boolean Sync() {
        Log.v(TAG,"Starting sync");
        String server = this.GetKey("server");
        String devicename = this.GetKey("devicename");
        String apikey = this.GetKey("__apikey");

        ConnectorDB cdb;
        try {
            cdb=new ConnectorDB(server,devicename,apikey);
        } catch (IllegalStateException e) {
            Log.e(TAG,"Could not initialize connectordb");
            return false;
        }

        if (!cdb.Ping()) {
            Log.e(TAG,"Ping failed");
            return false;
        }

        SQLiteDatabase db = this.getWritableDatabase();

        //For each stream in database
        Cursor res = db.rawQuery("SELECT streamname,schema FROM streams", new String[]{});
        int resultcount = res.getCount();
        if (resultcount ==0 ) {
            Log.i(TAG,"No streams to sync");
            return true;
        }

        for (int i =0; i<resultcount; i++) {
            res.moveToNext();
            String streamname = res.getString(0);
            String schema = res.getString(1);
            Log.v(TAG,"Syncing stream "+streamname);
            if (!cdb.HasStream(streamname)) {
                Log.w(TAG,"Stream does not exist: "+streamname);
                //Create the stream
                if (!cdb.CreateStream(streamname,schema)) {
                    Log.e(TAG,"Creating stream failed: "+streamname);
                    return false;
                }
            }

            //Insert the datapoints
            Cursor dta = db.rawQuery("SELECT timestamp,data FROM cache WHERE streamname=? ORDER BY timestamp ASC;", new String[]{streamname});
            int dtacount = dta.getCount();

            Log.i(TAG,"Writing "+dtacount+" datapoints to "+streamname);

            //Get the most recently inserted timestamp
            double oldtime = 0;
            String keyname = "sync_oldtime_"+streamname;
            try {
                oldtime = Double.parseDouble(GetKey(keyname));
            } catch(NumberFormatException nfe) {}

            StringBuilder totaldata = new StringBuilder();
            totaldata.append("[");
            for (int j=0; j< dtacount; j++) {
                dta.moveToNext();
                double timestamp = dta.getDouble(0);
                if (timestamp>oldtime) {
                    oldtime = timestamp;
                    totaldata.append("{\"t\": ");
                    totaldata.append(timestamp);
                    totaldata.append(", \"d\": ");
                    totaldata.append(dta.getString(1));
                    totaldata.append("},");
                } else {
                    Log.w(TAG,streamname+": Skipping duplicate timestamp");
                }
            }
            String totaldatas = totaldata.toString();
            totaldatas = totaldatas.substring(0, totaldata.length()-1)+"]";

            if (totaldatas.length()>1) {
                if (!cdb.Insert(streamname,totaldatas)) {
                    Log.e(TAG,"FAILED TO INSERT "+streamname);
                    return false;
                }

                //Now delete the data from the cache
                db.execSQL("DELETE FROM cache WHERE streamname=? AND timestamp <=?",new Object[]{streamname, oldtime});

                SetKey(keyname, Double.toString(oldtime));
            }

        }

        Log.v(TAG,"Sync successful");
        return true;
    }


    public String GetKey(String key) {
        SQLiteDatabase db = this.getReadableDatabase();
        Cursor res = db.rawQuery("SELECT value FROM kv WHERE key=?;", new String[]{key});
        if (res.getCount() ==0 ) {
            return "";
        } else {
            res.moveToNext();
            if (key.startsWith("__")) {
                Log.v(TAG, "Got: *****");
            } else {
                Log.v(TAG, "Got: " + res.getString(0));
            }
            return res.getString(0);
        }
    }
    public void SetKey(String key,String value) {
        if (key.startsWith("__")) {
            Log.v(TAG, "SET " + key + " TO ********");
        }else{
            Log.v(TAG, "SET " + key + " TO " + value);
        }
        SQLiteDatabase db = this.getWritableDatabase();
        ContentValues contentValues = new ContentValues();
        contentValues.put("key", key);
        contentValues.put("value", value);
        db.replace("kv",null,contentValues);
    }

    public void Clear() {
        SQLiteDatabase db = this.getWritableDatabase();
        db.execSQL("DELETE FROM cache;");
    }

    public void SetCred(String device, String apikey) {
        this.SetKey("devicename",device);
        this.SetKey("__apikey",apikey);
    }
}