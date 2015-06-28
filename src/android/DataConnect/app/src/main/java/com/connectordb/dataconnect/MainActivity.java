package com.connectordb.dataconnect;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;

import org.apache.cordova.*;

public class MainActivity extends CordovaActivity
{
    private ApiInitializer apiInitializer;
    @Override
    public void onCreate(Bundle savedInstanceState)
    {
        super.onCreate(savedInstanceState);

        // Set by <content src="index.html" /> in config.xml
        loadUrl(launchUrl);

        //Makes sure that we have permissions to use all the requested APIs
        apiInitializer = new ApiInitializer(this);

        // Start the logger service if it isn't running
        startService(new Intent(this,LoggerService.class));

    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        if (requestCode == 1) {
            if (resultCode == RESULT_OK) {
                apiInitializer.reconnect();
            }
        } else {
            super.onActivityResult(requestCode,resultCode,data);
        }
    }
}
