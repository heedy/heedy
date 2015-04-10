package com.connectordb.dataconnect;

import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.DatabaseUtils;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;
import android.util.Log;

/**
 * Created by Daniel on 2/22/2015.
 */
public class DataCache extends SQLiteOpenHelper {
    public static final int DATABASE_VERSION = 3;   //Version 2: ConnectorDB 0.1 had different schema types - so want to purge data
    private static final String TAG = "DataCache";
    public static final String DATABASE_NAME = "DataCache.db";

    //Make this shit a singleton
    private static DataCache dc;
    public static synchronized DataCache get(Context c) {
        if (dc==null) {
            Log.v(TAG,"Initializing DataCache");
            dc = new DataCache(c);
        }
        return dc;
    }


    public DataCache(Context context)
    {
        super(context, DATABASE_NAME , null, DATABASE_VERSION);
    }

    @Override
    public void onCreate(SQLiteDatabase db) {
        // TODO Auto-generated method stub
        db.execSQL(
                "create table datacache " +
                        "(id integer primary key, timestamp integer, stream text, data text )"
        );
        db.execSQL(
                "create table kvstore " +
                        "(key text primary key, value text)"
        );
    }

    @Override
    public void onUpgrade(SQLiteDatabase db, int oldVersion, int newVersion) {
        // TODO Auto-generated method stub
        Log.w(TAG,"Upgrading DataCache - deleting cached table");
        db.execSQL("DROP TABLE IF EXISTS datacache");
        onCreate(db);
    }

    public Cursor getCache(){
        SQLiteDatabase db = this.getReadableDatabase();
        Cursor res =  db.rawQuery( "SELECT * FROM datacache ORDER BY timestamp ASC;", null );
        return res;
    }

    public int Size(){
        SQLiteDatabase db = this.getReadableDatabase();
        int numRows = (int) DatabaseUtils.queryNumEntries(db, "datacache");
        Log.v(TAG,"Cache Size: "+ Integer.toString(numRows));
        return numRows;
    }

    public void Insert(String stream, long timestamp, String data) {
            Log.v(TAG,"[s=" + stream + " t=" + Long.toString(timestamp) + " d=" + data + "]");
            SQLiteDatabase db = this.getWritableDatabase();
            ContentValues contentValues = new ContentValues();
            contentValues.put("stream", stream);
            contentValues.put("timestamp", timestamp);
            contentValues.put("data", data);
            db.insert("datacache",null,contentValues);
    }

    public Boolean TExists(String stream, long timestamp) {
        SQLiteDatabase db = this.getReadableDatabase();
        Cursor res =  db.rawQuery( "SELECT timestamp FROM datacache WHERE stream=? AND timestamp>=?;", new String[] {stream,Long.toString(timestamp)});
        return res.getCount()>0;
    }

    public void Delete(long id) {
        SQLiteDatabase db = this.getWritableDatabase();
        db.delete("datacache","id=?",new String[] {Long.toString(id)});
    }

    public String GetKey(String key) {
        SQLiteDatabase db = this.getReadableDatabase();
        Cursor res = db.rawQuery( "SELECT value FROM kvstore WHERE key=?;", new String[] {key});
        if (res.getCount() ==0 ) {
            return "";
        } else {
            res.moveToNext();
            return res.getString(0);
        }
    }
    public void SetKey(String key,String value) {
        SQLiteDatabase db = this.getReadableDatabase();
        Cursor res = db.rawQuery( "INSERT OR REPLACE INTO kvstore (key,value) VALUES (?,?);", new String[] {key,value});
    }

}
