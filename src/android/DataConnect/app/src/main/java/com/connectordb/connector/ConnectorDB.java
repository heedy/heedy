package com.connectordb.connector;


import android.util.Base64;
import android.util.Log;

import java.lang.annotation.Documented;
import java.lang.annotation.Retention;
import java.lang.annotation.Target;

import static java.lang.annotation.ElementType.METHOD;
import static java.lang.annotation.RetentionPolicy.RUNTIME;

import retrofit.RequestInterceptor;
import retrofit.RestAdapter;
import retrofit.RetrofitError;
import retrofit.http.Body;
import retrofit.http.GET;
import retrofit.http.POST;
import retrofit.http.Path;
import retrofit.http.RestMethod;

import retrofit.http.RestMethod;

public class ConnectorDB {
    public static final String TAG= "ConnectorDB:Connection";

    public final String user;
    public final String device;
    public final String apikey;

    //ConnectorDB needs the UPDATE http verb
    @Documented
    @Target(METHOD)
    @Retention(RUNTIME)
    @RestMethod(value = "UPDATE", hasBody = true)
    private @interface UPDATE {
        String value();
    }


    private interface CDBAPI {
        @GET("/api/v1/d/?q=this")
        String Ping();

    }

    private CDBAPI dbapi;

    ConnectorDB(String server, String devicename, String apikey) {
        String[] path = devicename.split("/");

        if (path.length!=2) {
            Log.e(TAG, "Device name bad:" + devicename);
            throw new IllegalStateException();
        }
        this.user = path[0];
        this.device = path[1];
        this.apikey = apikey;

        RequestInterceptor requestInterceptor = new RequestInterceptor() {
            @Override
            public void intercept(RequestInterceptor.RequestFacade request) {
                String credentials = ConnectorDB.this.device+":"+ConnectorDB.this.apikey;
                String basicauth = "Basic " + Base64.encodeToString(credentials.getBytes(), Base64.NO_WRAP);
                request.addHeader("Authorization", basicauth);
            }
        };
        RestAdapter restAdapter = new RestAdapter.Builder()
                .setEndpoint(server)
                .setRequestInterceptor(requestInterceptor)
                .build();
        this.dbapi = restAdapter.create(CDBAPI.class);
    }

    String Ping() {
        Log.v(TAG,"PING");
        try {
            return this.dbapi.Ping();
        } catch (RetrofitError e) {
            Log.e(TAG,"PING: "+e.toString());
            return "";
        }
    }


}
