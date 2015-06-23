package com.connectordb.dataconnect;


import android.util.Log;

import com.connectordb.connector.Logger;

import org.apache.cordova.CallbackContext;
import org.apache.cordova.CordovaPlugin;
import org.apache.cordova.PluginResult;
import org.json.JSONArray;
import org.json.JSONException;

public class DataConnectorPlugin extends CordovaPlugin {
    public static final String TAG = "DataConnectorPlugin";

    public DataConnectorPlugin() {}

    public boolean execute(final String action, JSONArray args, CallbackContext callbackContext) throws JSONException {
        if (action.equals("setcred")) {
            Log.i(TAG,"Setting sync credentials: "+args.getString(0)+" "+args.getString(1));
            Logger.get(this.webView.getContext()).SetCred(args.getString(0),args.getString(1));

            callbackContext.success();
            return true;
        } else if (action.equals("getcachesize")) {
            Log.v(TAG,"Getting cache size");
            //callbackContext.sendPluginResult(new PluginResult(PluginResult.Status.OK,Logger.get(this.webView.getContext()).Size()));
            callbackContext.success(Logger.get(this.webView.getContext()).Size());
            return true;
        } else if (action.equals("sync")) {
            Log.i(TAG,"Running sync");
            Logger.get(this.webView.getContext()).BGSync();
            callbackContext.success();
            return true;
        } else if (action.equals("clear")) {
            Log.i(TAG,"Clear cache");
            Logger.get(this.webView.getContext()).Clear();
            callbackContext.success();
            return true;
        } else if (action.equals("background")) {
            double synctime = args.getDouble(0);
            Log.i(TAG,"Setting background sync: "+synctime);
            if (synctime <= 1.) {
                //Disable bg sync
                Logger.get(this.webView.getContext()).DisableTimedSync();
            } else {
                //Set bg sync to the given number
                Logger.get(this.webView.getContext()).EnableTimedSync((long)(synctime*1000.));
            }
            callbackContext.success();
            return true;
        }
        return false;
    }
}
