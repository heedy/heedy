package com.connectordb.dataconnect;


import android.util.Log;

import org.apache.cordova.CallbackContext;
import org.apache.cordova.CordovaPlugin;
import org.json.JSONArray;
import org.json.JSONException;

public class DataConnectorPlugin extends CordovaPlugin {
    public static final String TAG = "DataConnectorPlugin";

    public DataConnectorPlugin() {}

    public boolean execute(final String action, JSONArray args, CallbackContext callbackContext) throws JSONException {
        Log.d(TAG,"execute called: "+action);
        if (action.equals("setcred")) {
            Log.i(TAG,"Setting sync credentials: "+args.getString(0)+" "+args.getString(1));
            callbackContext.success();
            return true;
        }
        return false;
    }
}
