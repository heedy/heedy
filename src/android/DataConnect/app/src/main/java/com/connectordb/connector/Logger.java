package com.connectordb.connector;

import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.DatabaseUtils;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;
import android.util.Log;

public class Logger extends SQLiteOpenHelper {
    public static final int DATABASE_VERSION = 1;
    public static final String TAG = "ConnectorDB:Logger";
    public static final String DATABASE_NAME = "ConnectorLog.db";

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
    }

    @Override
    public void onCreate(SQLiteDatabase db) {
        db.execSQL("CREATE TABLE streams (streamname TEXT PRIMARY KEY, schema TEXT);");
        db.execSQL("CREATE TABLE cache (streamname TEXT, timestamp REAL, data TEXT);");
        db.execSQL("CREATE TABLE kv (key TEXT PRIMARY KEY, value TEXT);");

        //Now fill in the default values in kv
        db.execSQL("INSERT INTO kv VALUES ('server','https://connectordb.com');");
        db.execSQL("INSERT INTO kv VALUES ('devicename','');");
        db.execSQL("INSERT INTO kv VALUES ('__apikey','');");
        db.execSQL("INSERT INTO kv VALUES ('syncperiod','3600');"); //Make the database sync every hour
        db.execSQL("INSERT INTO kv VALUES ('autorun','0');");   //Do not run automatically

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
        Log.v(TAG,"Ensuring stream "+stream);

        SQLiteDatabase db = this.getWritableDatabase();
        ContentValues contentValues = new ContentValues();
        contentValues.put("streamname", stream);
        contentValues.put("schema", schema);
        db.insertWithOnConflict("streams",null,contentValues,SQLiteDatabase.CONFLICT_REPLACE);
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

    /*
    //Synchronizes the database with the server
    public synchronized boolean Sync() {
        String server = this.GetKey("server");
        String devicename = this.GetKey("devicename");
        String apikey = this.GetKey("__apikey");

        ConnectorDB cdb;
        try {
            cdb=new ConnectorDB(server,devicename,apikey);
        } catch (IllegalStateException e) {
            return false;
        }

        if (cdb.Ping().isEmpty()) {
            return false;
        }

        SQLiteDatabase db = this.getWritableDatabase();


        return true;
    }
    */

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

    public void SetCred(String device, String apikey) {
        this.SetKey("devicename",device);
        this.SetKey("apikey",apikey);
    }
}