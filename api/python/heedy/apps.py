from typing import Dict
from .base import APIObject, APIList, HeedyException, Session, getSessionType, DEFAULT_URL, q
from .kv import KV

from . import users
from . import objects
from .notifications import Notifications

from functools import partial


class App(APIObject):
    """
    App is a class representing a Heedy app. Using an access token, you can log into heedy
    using the App object::

    .. tab:: Sync

        ::

            access_token = "..."
            app = heedy.App(access_token,url="http://localhost:1324")


    .. tab:: Async

        ::

            access_token = "..."
            app = heedy.App(access_token,url="http://localhost:1324",session="async")

        access_token = "..."
        app = heedy.App(access_token,url="http://localhost:1324")

    Pre-initialized app objects are also returned when querying for apps from a plugin.
    """

    props = {
        "name",
        "description",
        "icon",
        "settings",
        "settings_schema",
        "access_token",
    }

    read_qparams = {"icon": {"icon": True}, "access_token": {"token": True}}

    def __init__(
        self, access_token: str, url: str = DEFAULT_URL, session="sync", cached_data=None
    ):
        appid = "self"
        if isinstance(session, Session):
            # Treat the session as already initialized, meaning that the access token is actually
            # the app id
            appid = access_token
            super().__init__(
                f"api/apps/{q(appid)}", {"app": appid}, session, cached_data=cached_data
            )

        else:
            # Initialize the app object as a direct API
            s = getSessionType(session, "self", url)
            s.setAccessToken(access_token)
            super().__init__(
                "api/apps/self", {"app": appid}, s, cached_data=cached_data
            )

            # Cache the used access token
            self.cached_data["access_token"] = access_token

        # Key-value store associated with the app
        self._kv = KV(f"api/kv/apps/{q(appid)}", self.session)

        if not "id" in self.cached_data:
            self.cached_data["id"] = appid

        #: An :class:`~heedy.Objects` instance associated with the app, allowing
        #: to interact with the objects of the app. For example, listing the objects
        #: that are managed by the app can be done with ``app.objects()``.
        self.objects = objects.Objects({"app": appid}, self.session)

    @property
    def kv(self):
        """
        The key-value store associated with this app. For details of usage, see :ref:`python_kv`.

        Returns:
            A :class:`heedy.kv.KV` object for the element.
        """
        return self._kv

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)

    @property
    def id(self):
        """
        The app's unique ID. This is directly available, so does not need
        to be awaited in async sessions::

            print(myapp.id)
        """
        return self.cached_data["id"]

    def __getitem__(self, i):
        v = super().__getitem__(i)
        if i == "owner":
            return users.User(v, self.session)
        return v

    @property
    def owner(self):
        """
        The user which owns this app:

        .. tab:: Sync

            ::

                print(app.owner.username) # Queries the server for the owner, and prints username
                print(app["owner"].username) # Uses cached query data to get the owner


        .. tab:: Async

            ::

                print((await app.owner).username) # Queries the server for the owner, and prints username
                print(app["owner"].username) # Uses cached query data to get the owner



        Returns:
            The :class:`~heedy.User` object of the user which owns this object.
            The returned user does not have any cached data other than its username.
        """
        return self.session.f(
            self.read(), lambda x: users.User(x["owner"], self.session)
        )

    def update(self, **kwargs):
        if "access_token" in kwargs and isinstance(kwargs["access_token"], bool):
            kwargs["access_token"] = "generate" if kwargs["access_token"] else ""
        return super().update(**kwargs)

    def copy(self, session: str = None):
        """
        Creates a copy of the app object. This function is usually used to switch between
        sync and async sessions::

            app_sync = heedy.App(access_token,url="http://localhost:1324",session="sync")
            app_async = app_sync.copy("async")

        This function will fail if used on an app object returned from a plugin session.
        Only App login sessions can be copied.

        Args:
            session (str): The session type to use for the copy (either "sync" or "async").
                If not specified, the session type of the current app is used.

        Returns:
            A new :class:`~heedy.App` object.
        """
        if session is None:
            session = "async" if self.session.isasync else "sync"
        return App(
            self["access_token"],
            url=self.session.url,
            session=session,
            cached_data=self.cached_data.copy(),
        )

    def __eq__(self,other):
        if self.id=="self" or other.id=="self":
            raise AttributeError("App object was not read from the server, and cannot be compared. Call .read() first.")
        return self.id==other.id

class Apps(APIList):
    """
    Apps is a class implementing a list of apps. It is accessed as a property of users
    to allow querying apps belonging to a user, or as a property of a plugin
    object, to allow generic querying of apps. Listing apps is currently only possible
    when authenticated as a plugin, because an app only has access to itself,
    and cannot read any other apps.

    .. tab:: Sync

        ::

            myuser.apps() # list of apps belonging to myuser
            plugin.apps() # list of all apps in heedy

    .. tab:: Async

        ::

            await myuser.apps() # list of apps belonging to myuser
            await plugin.apps() # list of all apps in heedy
    """

    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/apps", constraints, session)

    def __getitem__(self, appId: str):
        """Gets an app by its ID. Each app in heedy has a unique string ID,
        which can then be used to access the app.
        The ID can be seen in the URL of the app's page in the frontend.

        .. tab:: Sync

            ::

                app = p.apps["d233rk43o6kkle43kl"]

        .. tab:: Async

            ::

                app = await p.apps["d233rk43o6kkle43kl"]

        Returns:
            The :class:`~heedy.App` with the given ID (or promise for the app)
        Throws:
            HeedyException: If the app does not exist, or insufficient permissions

        """
        return self._getitem(
            appId, f=lambda x: App(x["id"], session=self.session, cached_data=x)
        )

    def __call__(self, **kwargs):
        """Gets the apps matching the given constraints.
        If used as a property of a user, the apps are those belonging to the user.
        Otherwise, if used as a property of a plugin, apps for all users are returned.

        .. tab:: Sync

            ::

                # Get all plugins handled by myplugin, with myapptype subkey.
                # Include each app's access tokens in the response.
                applist = p.apps(plugin="myplugin:myapptype",token=True)

        .. tab:: Async

            ::

                # Get all plugins handled by myplugin, with myapptype subkey.
                # Include each app's access tokens in the response.
                applist = await p.apps(plugin="myplugin:myapptype",token=True)

        Args:
            owner (str): The username of the apps' owner. Set automatically when accessed using the apps property of a user.
            plugin (str): The plugin key of the apps. Allows querying for apps related to specific plugins.
            enabled (bool,null): Include only enabled/disabled apps (true/false)
            icon (bool,False): Whether to include the icons of the apps in the response.
            token (bool,False): Whether to include the access tokens of the apps in the response.
        Returns:
            A list of :class:`~heedy.App` matching the given constraints.
        Throws:
            HeedyException: If the request fails.
        """
        return self._call(
            f=lambda x: [
                App(xx["id"], session=self.session, cached_data=xx) for xx in x
            ],
            **kwargs,
        )

    def create(self, name: str="", **kwargs):
        """
        Creates a new app. Only the first argument, the app name, is required.

        .. tab:: Sync

            ::

                app = p.apps.create("My App",
                        description="This is my plugin's app",
                        icon="fas fa-chart-line",
                        owner="myuser", # set automatically when accessed using the apps property of a user
                        plugin="myplugin:myapptype"
                    )

        .. tab:: Async

            ::

                app = await p.apps.create("My App",
                        description="This is my plugin's app",
                        icon="fas fa-chart-line",
                        owner="myuser", # set automatically when accessed using the apps property of a user
                        plugin="myplugin:myapptype"
                    )

        When creating an app for a plugin, it is useful to set the :code:`plugin`
        property in the format :code:`plugin_name:app_type`. The plugin property
        allows a plugin to recover which apps are associated with it, and the subkey
        allows the plugin to distinguish between multiple apps with different functionality
        that are handled by the same plugin.

        Args:
            name (str,""): The name of the app.
            description (str,""): A description of the app.
            icon (str,""): The app's icon, either a base64 urlencoded image or fontawesome/material icon id.
            owner (str): The username of the app's owner. Set automatically when accessed using the apps property of a user.
            plugin (str): The plugin key of the app. Usually in the format :code:`plugin_name:app_type`.
            enabled (bool,True): Whether the app is enabled.
            scope (str,""): space-separated access scopes that are given to the app. Only relevant if the app will be accessed using its access token.
            settings (dict,{}): A dictionary of settings for the app.
            settings_schema (dict,{}): A JSON Schema describing the settings.
            access_token (bool,True): Whether the app should be given an access token.
        Returns:
            The newly created :class:`~heedy.App`, with its data cached.
        Throws:
            HeedyException: If the request fails.
        """
        if "access_token" in kwargs and isinstance(kwargs["access_token"], bool):
            kwargs["access_token"] = "generate" if kwargs["access_token"] else ""
        if name!="": # An empty name is allowed if creating a plugin app that is already defined in config
            kwargs["name"] = name
        return self._create(
            f=lambda x: App(x["id"], session=self.session, cached_data=x),
            **kwargs,
        )
