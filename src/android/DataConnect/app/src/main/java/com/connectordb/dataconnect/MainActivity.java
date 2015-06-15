package com.connectordb.dataconnect;

import android.content.Intent;
import android.os.Bundle;
import android.util.Log;

import org.apache.cordova.*;

public class MainActivity extends CordovaActivity
{
    @Override
    public void onCreate(Bundle savedInstanceState)
    {
        super.onCreate(savedInstanceState);

        // Set by <content src="index.html" /> in config.xml
        loadUrl(launchUrl);

        // Start the logger service if it isn't running
        startService(new Intent(this,LoggerService.class));

    }
}
