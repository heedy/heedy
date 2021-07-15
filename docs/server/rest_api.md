# REST API

Heedy's REST API allows programmatic access to backend data and functionality. Heedy plugins are also encouraged to add their own functionality at `/api/{plugin_name}`.

This document details the API that comes built-in to heedy by default.

## Authorization

Since heedy was built to be internet-facing, most resources are only available to authorized users.
The API can be accessed in three separate ways:

- By apps, with an app token. Apps have scoped access to the user's data.
- By plugins, using their plugin key. Plugins get full access to the API for all users.
- By users, using a browser cookie. This method is only used for the frontend.

Each of these access methods is described individually below.

```eval_rst
.. warning::
    Because access credentials are sent directly with each request, it is important to `use https <./installing.html#putting-heedy-online>`_ to secure
    any internet-accessible heedy instance.
```

### App Token

An app token requires each request to include an `Authorization` header:

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     http://localhost:1324/api/users/myuser
```

This method is used for all external heedy apps, and is limited in access to the scopes set for the app. You can get an app's access token in the app's page.

### Plugin Key

A backend plugin is given a plugin key in the json bundle passed to its stdin on startup (see [plugin backends](../plugins/backend/index.md)). It uses this key for all requests. The key is passed in a special `X-Heedy-Key` header:

```bash
curl --header "X-Heedy-Key: MYKEY" \
     http://localhost:1324/api/users/myuser
```

A plugin can access the API as if it were any user or app by including the `X-Heedy-As` header along with its plugin key (see [plugin backends](../plugins/backend/index.md)).

### User Cookie

When logged into heedy, a cookie is created in the browser which gives user-level access to the cookie's holder.
This means that once you log into heedy, you can use your browser to run GET api calls as your user by simply navigating to the api location.

All requests done from a plugin's frontend javascript module automatically include this cookie.

## Errors

Each request returns either the requested resource as JSON, or, upon failure, returns a `4xx` error code,
with the following json response body:

<div class="rest_output_result">

```javascript
{
  // An error type
  "error": "not_found",
  // Text description of the error
  "error_description": "The given user was not found",
  // ID of the request
  "id": "bpqirkjqfqj3m0sqttf0"
}
```

</div>

## API

### Users

<h4 class="rest_path">/api/users</h4>
<h5 class="rest_verb">GET</h5>
Returns a list of users in the heedy instance that are accessible with the given access token.

<h6 class="rest_params">URL Params</h6>

- **icon** _(boolean,false)_ - whether or not to include each user's icon.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     http://localhost:1324/api/users
```

<div class="rest_output_result">

```javascript
[{"username": "myuser", ... }, ... ]
```

</div>

<h5 class="rest_verb">POST</h5>
Create a new user. Only accessible from plugins and admin users.

<h6 class="rest_body">Body</h6>

- **username** _(string, required)_ - the username of the user
- **password** _(string, required)_ - the user's password
- **name** _(string,"")_ - the user's full name
- **description** _(string,"")_ - the user's description
- **icon** _(string,"")_ - user's icon, base64 urlencoded
- **public_read** _(boolean,false)_ - whether the user is visible to the public.
- **users_read** _(boolean,false)_ - whether the user is visible to other users.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --header "Content-Type: application/json" \
     --request POST \
     --data '{"username":"user2","password":"xyz"}' \
     http://localhost:1324/api/users
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

<h4 class="rest_path">/api/users/<span>{username}</span></h4>
<h5 class="rest_verb">GET</h5>
Returns the user with the given username.
<h6 class="rest_params">URL Params</h6>

- **icon** _(boolean,false)_ - whether or not to include the user's icon.

<h6 class="rest_output">Example</h6>
```bash
curl --header "Authorization: Bearer MYTOKEN" \
     http://localhost:1324/api/users/myuser?icon=true
```

<div class="rest_output_result">

```javascript
{
    "username":"myuser",
    "name":"",
    "description":"",
    "icon":"",
    "public_read":false,
    "users_read":false
}
```

</div>

<h5 class="rest_verb">PATCH</h5>
Updates the user with the included fields.

<h6 class="rest_body">Body</h6>

- **password** _(string,null)_ - the user's password
- **name** _(string,null)_ - the user's full name
- **description** _(string,null)_ - the user's description
- **icon** _(string,null)_ - user's icon, base64 urlencoded
- **public_read** _(boolean,null)_ - whether the user is visible to the public.
- **users_read** _(boolean,null)_ - whether the user is visible to other users.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --header "Content-Type: application/json" \
     --request PATCH \
     --data '{"description":"A new user description!"}' \
     http://localhost:1324/api/users/myuser
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

<h5 class="rest_verb">DELETE</h5>
Deletes the user with the given username, and all of the user's data.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request DELETE \
     http://localhost:1324/api/users/myuser
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

### Apps

<h4 class="rest_path">/api/apps</h4>
<h5 class="rest_verb">GET</h5>
Returns a list of apps in the heedy instance that satisfy the given constraints, and are accessible with the given access token. A user only has access to apps belonging to the user, so when querying apps from the frontend or an app, you will only get the user's apps.

<h6 class="rest_params">URL Params</h6>

- **icon** _(boolean,false)_ - whether or not to include each app's icon.
- **token** _(boolean,false)_ - whether or not to include each app's access token.
- **owner** _(string,null)_ - limit results to the apps belonging to the given username
- **plugin** _(string,null)_ - limit results to apps with the given plugin key

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     http://localhost:1324/api/apps?owner=myuser
```

<div class="rest_output_result">

```javascript
[{"id": "051942...", ... }, ... ]
```

</div>

<h5 class="rest_verb">POST</h5>
Create a new app. Only accessible from plugins and users. If authenticated as a user, the owner parameter is optional.

<h6 class="rest_body">Body</h6>

- **name** _(string,required)_ - the app's name
- **owner** _(string,required)_ - the user to own the app. Automatically set to the current user when authenticated as a user.
- **description** _(string,"")_ - the app's description
- **icon** _(string,"")_ - app's icon, base64 urlencoded
- **plugin** _(string,"")_ - the app's plugin key.
- **enabled** _(boolean,true)_ - whether the app's access token is active
- **scope** _(string,"")_ - the scopes given to the app, each separated by a space.
- **settings** _(object,{})_ - the app's settings
- **settings_schema** _(object,{})_ - the json schema for the app's settings.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --header "Content-Type: application/json" \
     --request POST \
     --data '{"name":"My App","owner": "myuser"}' \
     http://localhost:1324/api/apps
```

<div class="rest_output_result">

```javascript
{
    "id":"0519420b-e3cf-463f-b794-2adb440bfb9f",
    "name":"My App",
    "description":"",
    "owner":"myuser",
    "enabled":true,
    "created_date":"2020-03-21",
    "last_access_date":null,
    "scope":"",
    "settings":{},
    "settings_schema":{}
}
```

</div>

<h4 class="rest_path">/api/apps/<span>{appid}</span></h4>
<h5 class="rest_verb">GET</h5>
Returns the app with the given ID
<h6 class="rest_params">URL Params</h6>

- **icon** _(boolean,false)_ - whether or not to include the app's icon.
- **token** _(boolean,false)_ - whether or not to include the app's access token.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
 http://localhost:1324/api/apps/0519420b-e3cf-463f-b794-2adb440bfb9f
```

<div class="rest_output_result">

```javascript
{
    "id": "0519420b-e3cf-463f-b794-2adb440bfb9f",
    "name": "My App",
    "description": "",
    "owner": "myuser",
    "enabled": true,
    "created_date": "2020-03-21",
    "last_access_date": null,
    "scope": "",
    "settings": {},
    "settings_schema": {}
}
```

</div>

<h5 class="rest_verb">PATCH</h5>
Updates the app with the included fields.

<h6 class="rest_body">Body</h6>

- **name** _(string,null)_ - the app's full name
- **description** _(string,null)_ - the app's description
- **icon** _(string,null)_ - app's icon, base64 urlencoded
- **plugin** _(string,null)_ - the app's plugin key.
- **enabled** _(boolean,null)_ - whether the app's access token is active
- **scope** _(string,null)_ - the scopes given to the app, each separated by a space.
- **settings** _(object,null)_ - the app's settings
- **settings_schema** _(object,null)_ - the json schema for the app's settings.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --header "Content-Type: application/json" \
     --request PATCH \
     --data '{"description":"A new app description!"}' \
 http://localhost:1324/api/apps/0519420b-e3cf-463f-b794-2adb440bfb9f
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

<h5 class="rest_verb">DELETE</h5>
Deletes the given app, and all of its data, including objects it manages.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request DELETE \
 http://localhost:1324/api/apps/0519420b-e3cf-463f-b794-2adb440bfb9f
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

### Objects

Heedy objects are special, since each object type has its own API. This section first describes the general object API that is valid for all object types, then it describes the additional API for objects of the type timeseries.

<h4 class="rest_path">/api/objects</h4>
<h5 class="rest_verb">GET</h5>
Returns a list of objects in the heedy instance satisfying the given constraints, and accessible to the authenticated entity.

<h6 class="rest_params">URL Params</h6>

- **icon** _(boolean,false)_ - whether or not to include each object's icon.
- **owner** _(string,null)_ - limit results to the objects belonging to the given username
- **app** _(string,null)_ - limit results to objects belonging to the given app
- **key** _(string,null)_ - limit results to objects with the given key
- **tags** _(string,null)_ - limit results to objects which each include _all_ the given tags
- **type** _(string,null)_ - limit results to objects of the given type
- **limit** _(int,null)_ - set a maximum number of results to return

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     http://localhost:1324/api/objects?owner=myuser
```

<div class="rest_output_result">

```javascript
[{"id": "1a1f624...", ... }, ... ]
```

</div>

<h5 class="rest_verb">POST</h5>
Create a new object of the given type. Unless owner/app is set, the object will belong to the authenticated entity.

<h6 class="rest_body">Body</h6>

- **name** _(string,required)_ - the object's name
- **type** _(string,required)_ - the object's type
- **owner** _(string,current_user)_ - the user to own the object.
- **app** _(string,current_app/null)_ - the app to own the object.
- **description** _(string,"")_ - the object's description
- **icon** _(string,"")_ - object's icon, base64 urlencoded
- **tags** _(string,"")_ - a set of space-separated tags to give the object
- **key** _(string,null)_ - a key to give the object for easy programmatic access (only objects belonging to apps can have keys)
- **owner_scope** _(string,"*")_ - the set of space-separated scopes to give the object's owner, if the object belongs to an app. "*" means all scopes.
- **meta** _(object,{})_ - object metadata. Each object type defines its own metadata.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --header "Content-Type: application/json" \
     --request POST \
     --data '{"name":"My Timeseries","type":"timeseries"}' \
     http://localhost:1324/api/objects
```

<div class="rest_output_result">

```javascript
{
    "id":"1a1f624e-96f9-416a-9982-6b1ef618661c",
    "name":"My TS",
    "description":"",
    "owner":"myuser",
    "app": "0519420b-e3cf-463f-b794-2adb440bfb9f",
    "tags":"",
    "type":"timeseries",
    "meta":{"actor":false,"schema":{}},
    "created_date":"2020-03-21",
    "modified_date":null,
    "owner_scope":"*",
    "access":"*"
}
```

</div>

<h4 class="rest_path">/api/objects/<span>{objectid}</span></h4>
<h5 class="rest_verb">GET</h5>
Returns the object with the given ID
<h6 class="rest_params">URL Params</h6>

- **icon** _(boolean,false)_ - whether or not to include the object's icon.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c
```

<div class="rest_output_result">

```javascript
{
    "id":"1a1f624e-96f9-416a-9982-6b1ef618661c",
    "name":"My TS",
    "description":"",
    "owner":"myuser",
    "app": "0519420b-e3cf-463f-b794-2adb440bfb9f",
    "tags":"",
    "type":"timeseries",
    "meta":{"actor":false,"schema":{}},
    "created_date":"2020-03-21",
    "modified_date":null,
    "owner_scope":"*",
    "access":"*"
}
```

</div>

<h5 class="rest_verb">PATCH</h5>
Updates the object with the included fields.

The `meta` object is updated on a per-field basis, meaning that
the object sent as the meta field will be merged with the existing meta values. Setting `meta` to `{"schema":{"type":"number"}}` in a timeseries object will update the `schema` of the meta object, leaving all other fields intact. To delete a field from the meta object, set it to null (`{"actor":null}`).

<h6 class="rest_body">Body</h6>

- **name** _(string,null)_ - the object's name
- **description** _(string,null)_ - the object's description
- **icon** _(string,null)_ - object's icon, base64 urlencoded
- **tags** _(string,null)_ - a set of space-separated tags to give the object
- **key** _(string,null)_ - a key to give the object for easy programmatic access (only objects belonging to apps can have keys)
- **owner_scope** _(string,null)_ - the set of space-separated scopes to give the object's owner, if the object belongs to an app. "*" means all scopes.
- **meta** _(object,null)_ - the fields of object metadata to update. Each object type defines its own metadata.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --header "Content-Type: application/json" \
     --request PATCH \
     --data '{"meta":{"schema":{"type":"number"}}}' \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

<h5 class="rest_verb">DELETE</h5>
Deletes the given object, and all of its data.

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --request DELETE \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c
```

<div class="rest_output_result">

```javascript
{"result":"ok"}
```

</div>

#### Timeseries

The timeseries is a builtin object type. It defines its own API for interacting with the datapoints contained in the series.

##### Meta

Each object holds a `meta` field. A timeseries object's meta object has the following fields:

- **schema** _(object,{})_ - a [JSON Schema](https://json-schema.org/) to which each datapoint must conform.

<h4 class="rest_path">/api/objects/<span>{objectid}</span>/timeseries</h4>
<h5 class="rest_verb">GET</h5>
Returns the timeseries data subject to the given constraints.
<h6 class="rest_params">URL Params</h6>

- **t** _(float,null)_ - get just the datapoint with the given timestamp
- **i** _(int,null)_ - get just the datapoint at the given index
- **t1** _(float/string\*,null)_ - return only datapoints where `t >= t1`
- **t2** _(float/string\*,null)_ - return only datapoints where `t < t2`
- **i1** _(int,null)_ - return only datapoints where `index >= i1`
- **i2** _(int,null)_ - return only datapoints where `index < i2`
- **limit** _(int,null)_ - return a maximum of this number of datapoints
- **transform** _(string,null)_ - a [PipeScript](/analysis/pipescript) transform to run on the data

_\*: The `t1` and `t2` queries accept strings of times relative to now. For example, `t1=now-2d` sets `t1` to exactly 2 days ago._

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c/timeseries?t1=now-2h
```

<div class="rest_output_result">

```json
[
  { "t": 1584812297, "d": 3 },
  { "t": 1584812303, "d": 2 },
  { "t": 1584812313, "d": 2 },
  { "t": 1584812339, "d": 2 }
]
```

</div>

<h5 class="rest_verb">POST</h5>
Insert new datapoints into the timeseries
<h6 class="rest_params">URL Params</h6>

- **method** _(string,"update")_ - the insert method, one of:
  - update - _Overwrite any datapoints that already exist with time ranges defined by the inserted datapoints._
  - append - _Only permit appending datapoints to the end of the timeseries_
  - insert - _Don't permit inserting datapoints that interfere with data already in the timeseries_

<h6 class="rest_body">Body</h6>
A json array of datapoints, conforming to the timeseries schema, with each datapoint in the following format:

```javascript
{
    // unix timestamp in seconds
    "t": 1584812297.1,
    // (optional) duration of the datapoint
    "dt": 60.0,
    // the datapoint's data (anything that can be encoded as json)
    "d": 3
}
```

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --header "Content-Type: application/json" \
     --request POST \
     --data '[{"t":1584812297,"d":3},{"t":1584812303,"d":2}]' \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c/timeseries
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h5 class="rest_verb">DELETE</h5>
Delete the timeseries data that satisfies the given constraints
<h6 class="rest_params">URL Params</h6>

- **t** _(float,null)_ - remove just the datapoint with the given timestamp
- **i** _(int,null)_ - remove just the datapoint at the given index
- **t1** _(float/string\*,null)_ - remove only datapoints where `t >= t1`
- **t2** _(float/string\*,null)_ - remove only datapoints where `t < t2`
- **i1** _(int,null)_ - remove only datapoints where `index >= i1`
- **i2** _(int,null)_ - remove only datapoints where `index < i2`

_\*: The `t1` and `t2` queries accept strings of times relative to now. For example, `t1=now-2d` sets `t1` to exactly 2 days ago._

<h6 class="rest_output">Example</h6>

```bash
curl --header "Authorization: Bearer MYTOKEN" \
     --request DELETE \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c/timeseries?t1=now-2h
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h4 class="rest_path">/api/objects/<span>{objectid}</span>/timeseries/length</h4>
<h5 class="rest_verb">GET</h5>
Returns the number of datapoints in the timeseries.

```bash
curl --header "Authorization: Bearer MYTOKEN" \
 http://localhost:1324/api/objects/1a1f624e-96f9-416a-9982-6b1ef618661c/timeseries/length
```

<div class="rest_output_result">

```json
4
```

</div>

### Notifications

Notifications are a built-in plugin that allows attaching messages to users/apps/objects. These messages are visible from the main heedy UI.

<h4 class="rest_path">/api/notifications</h4>
<h5 class="rest_verb">GET</h5>
Read the list of notifications subject to the given constraints.
<h6 class="rest_params">URL Params</h6>

- **key** _(string,null)_ - only return the notifications with the given key
- **user** _(string,null)_ - limit to notifications for the given user
- **app** _(string,null)_ - limit to notifications for the given app
- **object** _(string,null)_ - limit to notifications for the given app
- **global** _(boolean,false)_ - limit to notifications that show in the notifications page
- **seen** _(boolean,null)_ - limit to notifications that have/have not been seen
- **dismissible** _(boolean,null)_ - limit to notifications that are/are not dismissible
- **type** _(string,null)_ - limit to notifications of the given type
- **include_self** \_(boolean,false) - whether to include self when `*` present. For example, when `user=myuser&app=*`, notifications for user myuser are included if and only if `include_self` is true.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
      http://localhost:1324/api/notifications?user=myuser
```

<h5 class="rest_verb">POST</h5>

Create a notification for a user/app/object. If a notification with the given key exists for the given user/app/object, update the notification with the given data.

<h6 class="rest_body">Body</h6>

- **key** _(string,required)_ - set the notification's key (unique for the user/app/object's notifications)
- **title** _(string,"")_ - the header text to show
- **description** _(string,"")_ - main notification content. Can include markdown.
- **user** _(string,required*)_ - add the notification to the given username
- **app** _(string,required*)_ - add the notification to the given app
- **object** _(string,required*)_ - add the notification to the given object
- **global** _(boolean,false)_ - show in the global notification page?
- **seen** _(boolean,false)_ - has the notification been seen by the user?
- **dismissible** _(boolean,true)_ - allow the user to dismiss the notifcation
- **type** _(string,null)_ - the notification type, one of `info,warning,error`
- **actions** _(array,[])_ - the list of actions to give the notification, which are shown to the user as buttons. Each action object has the following fields:
    - **title** _(string,required)_ - the text to display in the button
    - **href** _(string,required)_ - the url to navigate to. If it starts with `#`, it is relative to the UI. If starts with `/`, relative to heedy's root. Otherwise, it is considered a raw URL.
    - **description** _(string,"")_ - the tooltip to show on button hover
    - **icon** _(string,"")_ - the icon to show in the button
    - **new_window** _(boolean,false)_ - whether to open href in a new window
    - **dismiss** _(boolean,false)_ - whether to dismiss the notification on click

_\*: Only one of the user/app/object fields can be set (the notification can only belong to a user or an app, or an object, not all at the same time)_

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --header "Content-Type: application/json" \
     --request POST \
     --data '{"key":"mynotification","user":"myuser","title": "Hello World!"}' \
     http://localhost:1324/api/notifications
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h5 class="rest_verb">PATCH</h5>

Modify the included fields of all notifications that satisfy the constraints given in the URL Params.

<h6 class="rest_params">URL Params</h6>

- **key** _(string,null)_ - only return the notifications with the given key
- **user** _(string,null)_ - limit to notifications for the given user
- **app** _(string,null)_ - limit to notifications for the given app
- **object** _(string,null)_ - limit to notifications for the given app
- **global** _(boolean,false)_ - limit to notifications that show in the notifications page
- **seen** _(boolean,null)_ - limit to notifications that have/have not been seen
- **dismissible** _(boolean,null)_ - limit to notifications that are/are not dismissible
- **type** _(string,null)_ - limit to notifications of the given type
- **include_self** \_(boolean,false) - whether to include self when `*` present. For example, when `user=myuser&app=*`, notifications for user myuser are included if and only if `include_self` is true.

<h6 class="rest_body">Body</h6>

- **key** _(string,null)_ - set the notification's key (unique for the user/app/object's notifications)
- **title** _(string,null)_ - the header text to show
- **description** _(string,null)_ - main notification content. Can include markdown.
- **global** _(boolean,null)_ - show in the global notification page?
- **seen** _(boolean,null)_ - has the notification been seen by the user?
- **dismissible** _(boolean,null)_ - allow the user to dismiss the notifcation
- **type** _(string,null)_ - the notification type, one of `info,warning,error`
- **actions** _(array,null)_ - the list of actions to give the notification, which are shown to the user as buttons. Each action object has the following fields:
    - **title** _(string,required)_ - the text to display in the button
    - **href** _(string,required)_ - the url to navigate to. If it starts with `#`, it is relative to the UI. If starts with `/`, relative to heedy's root. Otherwise, it is considered a raw URL.
    - **description** _(string,"")_ - the tooltip to show on button hover
    - **icon** _(string,"")_ - the icon to show in the button
    - **new_window** _(boolean,false)_ - whether to open href in a new window
    - **dismiss** _(boolean,false)_ - whether to dismiss the notification on click

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --header "Content-Type: application/json" \
     --request PATCH \
     --data '{"seen":true}' \
     http://localhost:1324/api/notifications?user=myuser&key=mynotification
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h5 class="rest_verb">DELETE</h5>

Deletes the notifications that satisfy the given constraints.

<h6 class="rest_params">URL Params</h6>

- **key** _(string,null)_ - delete the notifications with the given key
- **user** _(string,null)_ - limit to notifications for the given user
- **app** _(string,null)_ - limit to notifications for the given app
- **object** _(string,null)_ - limit to notifications for the given app
- **global** _(boolean,false)_ - limit to notifications that show in the notifications page
- **seen** _(boolean,null)_ - limit to notifications that have/have not been seen
- **dismissible** _(boolean,null)_ - limit to notifications that are/are not dismissible
- **type** _(string,null)_ - limit to notifications of the given type
- **include_self** \_(boolean,false) - whether to include self when `*` present. For example, when `user=myuser&app=*`, notifications for user myuser are included if and only if `include_self` is true.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request DELETE \
     http://localhost:1324/api/notifications?user=myuser&seen=true
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

### Key-Value Storage

The key-value database is a built-in plugin, allowing other plugins to store metadata attached to users, apps and objects. It is recommended that a plugin use its own plugin name as the namespace under which it stores its data.

An app can also store its own metadata, by using its app ID or `self` as the namespace when accessing the app's key-value store.

<h4 class="rest_path">/api/kv/users/<span>{id}</span>/<span>{namespace}</span></h4>
<h5 class="rest_verb">GET</h5>
Returns a json object containing all of the key-value pairs in the given namespace

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     http://localhost:1324/api/kv/users/myuser/myplugin
```

<div class="rest_output_result">

```json
{
  "mykey": 45.54
}
```

</div>

<h5 class="rest_verb">POST</h5>

Sets the key/values of the namespace to the posted body

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request POST \
     --header "Content-Type: application/json" \
     --data '{"mykey": 45.54}' \
     http://localhost:1324/api/kv/users/myuser/myplugin
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h5 class="rest_verb">PATCH</h5>

Updates only the given key/value pairs for the given namespace

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request PATCH \
     --header "Content-Type: application/json" \
     --data '{"mykey": 45.54}' \
     http://localhost:1324/api/kv/users/myuser/myplugin
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h4 class="rest_path">/api/kv/users/<span>{id}</span>/<span>{namespace}</span>/<span>{key}</span></h4>
<h5 class="rest_verb">GET</h5>
Get the value of the given key in the given namespace.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     http://localhost:1324/api/kv/users/myuser/myplugin/mykey
```

<div class="rest_output_result">

```json
45.54
```

</div>

<h5 class="rest_verb">POST</h5>

Sets the given key to the posted json value

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request POST \
     --header "Content-Type: application/json" \
     --data '45.54' \
     http://localhost:1324/api/kv/users/myuser/myplugin/mykey
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h5 class="rest_verb">DELETE</h5>

Deletes the given key from the given namespace.

<h6 class="rest_output">Example</h6>

```bash
curl --header "X-Heedy-Key: MYPLUGINKEY" \
     --request DELETE \
     http://localhost:1324/api/kv/users/myuser/myplugin/mykey
```

<div class="rest_output_result">

```json
{ "result": "ok" }
```

</div>

<h4 class="rest_path">/api/kv/apps/<span>{id}</span>/<span>{namespace}</span></h4>

Refer to `/api/kv/users/{id}/{namespace}`, which has an identical API

<h4 class="rest_path">/api/kv/apps/<span>{id}</span>/<span>{namespace}</span>/<span>{key}</span></h4>

Refer to `/api/kv/users/{id}/{namespace}/{key}`, which has an identical API

<h4 class="rest_path">/api/kv/objects/<span>{id}</span>/<span>{namespace}</span></h4>

Refer to `/api/kv/users/{id}/{namespace}`, which has an identical API

<h4 class="rest_path">/api/kv/objects/<span>{id}</span>/<span>{namespace}</span>/<span>{key}</span></h4>

Refer to `/api/kv/users/{id}/{namespace}/{key}`, which has an identical API
