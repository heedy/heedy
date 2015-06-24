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
import retrofit.client.ApacheClient;
import retrofit.http.Body;
import retrofit.http.GET;
import retrofit.client.Response;
import retrofit.http.POST;
import retrofit.http.Path;
import retrofit.http.RestMethod;

import retrofit.http.RestMethod;
import retrofit.mime.TypedByteArray;
import retrofit.mime.TypedInput;

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
        Response Ping();

        @GET("/api/v1/d/{user}/{device}/{stream}")
        Response GetStream(@Path("user") String user, @Path("device") String device, @Path("stream") String stream);

        @POST("/api/v1/d/{user}/{device}/{stream}")
        Response CreateStream(@Path("user") String user, @Path("device") String device, @Path("stream") String stream, @Body TypedInput obj);

        @UPDATE("/api/v1/d/{user}/{device}/{stream}")
        Response Insert(@Path("user") String user, @Path("device") String device, @Path("stream") String stream, @Body TypedInput obj);

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
                String credentials = ConnectorDB.this.user +"/"+ConnectorDB.this.device+":"+ConnectorDB.this.apikey;
                String basicauth = "Basic " + Base64.encodeToString(credentials.getBytes(), Base64.NO_WRAP);
                request.addHeader("Authorization", basicauth);
            }
        };
        RestAdapter restAdapter = new RestAdapter.Builder()
                .setEndpoint(server)
                .setRequestInterceptor(requestInterceptor)
                .setClient(new ApacheClient())
                .build();
        this.dbapi = restAdapter.create(CDBAPI.class);
    }

    boolean Ping() {
        Log.v(TAG,"PING");
        try {
            this.dbapi.Ping();
            return true;
        } catch (RetrofitError e) {
            Log.e(TAG,"PING: "+e.toString());
            return false;
        }
    }

    boolean HasStream(String sname) {
        Log.v(TAG,"Get stream "+sname);
        try {
            this.dbapi.GetStream(this.user, this.device, sname);
            return true;
        } catch (RetrofitError e) {
            Log.e(TAG,"GetStream: "+e.toString());
            return false;
        }
    }

    boolean CreateStream(String sname, String schema) {
        TypedInput in = new TypedByteArray("application/json", schema.getBytes());
        try {
            this.dbapi.CreateStream(this.user, this.device, sname, in);
            return true;
        } catch (RetrofitError e) {
            Log.e(TAG,"CreateStream: "+e.toString());
            return false;
        }
    }

    boolean Insert(String sname, String data) {
        TypedInput in = new TypedByteArray("application/json", data.getBytes());
        try {
            this.dbapi.Insert(this.user, this.device, sname, in);
            return true;
        } catch (RetrofitError e) {
            Log.e(TAG,"Insert: "+e.toString());
            return false;
        }
    }

}
