from typing import Dict
from .base import APIObject, APIList, Session, q
from .kv import KV

from . import apps
from . import objects
from .notifications import Notifications


class User(APIObject):
    props = {"name", "description", "icon", "public_read", "users_read"}

    def __init__(self, username: str, session: Session, cached_data=None):
        super().__init__(
            f"api/users/{q(username)}",
            {"user": username},
            session,
            cached_data=cached_data,
        )
        self.cached_data["username"] = username

        #: An :class:`~heedy.Apps` instance associated with the user, allowing
        #: to interact with the apps belonging to the user. For example, listing the
        #: user's apps can be done with ``user.apps()``.
        self.apps = apps.Apps({"owner": username}, self.session)

        #: An :class:`~heedy.Objects` instance associated with the user, allowing
        #: to interact with the user's objects. For example, listing the objects
        #: that are owned by the user can be done with ``usr.objects()``.
        self.objects = objects.Objects({"owner": username}, self.session)

        self._kv = KV(f"api/kv/users/{q(username)}", self.session)

    @property
    def kv(self):
        """
        The key-value store associated with this user. For details of usage, see :ref:`python_kv`.

        Returns:
            A :class:`heedy.kv.KV` object for the element.
        """
        return self._kv

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)

    @property
    def username(self):
        """
        The user's username. This is directly available, so does not need
        to be awaited in async sessions::

            print(myuser.username)
        """
        return self.cached_data["username"]

    @property
    def password(self):
        """
        The user's password is hashed in Heedy, and cannot be retrieved.
        However, a user's password can be changed by a plugin:

        .. tab:: Sync

            ::

                usr.password = "newpassword"

        .. tab:: Async

            ::

                await usr.update(password="newpassword")
        """
        raise PermissionError("A user's password is not readable")

    @password.setter
    def password(self, v):
        return self.update(password=v)

    def update(self, **kwargs):
        def remPass(x):
            # Remove the password from the updated cache
            self.cached_data.pop("password", None)
            return x
        return self.session.f(super().update(**kwargs), remPass)


class Users(APIList):
    """
    Users is a class implementing a list of users. It is accessed by plugins
    as a property of the :class:`heedy.Plugin` object. It is not currently possible to list
    users as an app.

    .. tab:: Sync

        ::

            plugin.users() # list of all users in heedy

    .. tab:: Async

        ::

            await plugin.users() # list of all users in heedy
    """

    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/users", constraints, session)

    def __getitem__(self, item):
        """Gets a user by their username.
        The username can be seen in the URL of the user's page in the frontend.

        .. tab:: Sync

            ::

                usr= p.users["myuser"]

        .. tab:: Async

            ::

                usr = await p.users["myuser"]

        Returns:
            The :class:`~heedy.User` with the given username (or promise for the user)
        Throws:
            HeedyException: If the user does not exist, or insufficient permissions

        """
        return self._getitem(
            item, f=lambda x: User(x["username"], session=self.session, cached_data=x)
        )

    def __call__(self, **kwargs):
        """
        Gets the list of users:

        .. tab:: Sync

            ::

                # list of all users in heedy, including their icons
                p.users(icon=True)

        .. tab:: Async

            ::

                # list of all users in heedy, including their icons
                await p.users(icon=True)

        Args:
            icon (bool,False): Whether to include the user's icon in the response.
        Returns:
            A list of :class:`~heedy.User` objects.
        Throws:
            HeedyException: If insufficient permissions or the request fails.
        """
        return self._call(
            f=lambda x: [
                User(xx["username"], session=self.session, cached_data=xx) for xx in x
            ],
            **kwargs,
        )

    def create(self, username, password, **kwargs):
        """
        Creates a new user with the given username and password.

        .. tab:: Sync

            ::

                usr = p.users.create("myusername", "mypassword",
                        name="Steve",
                        description="I like cake",
                        icon="person",
                        users_read=True,
                    )

        .. tab:: Async

            ::

                usr = await p.users.create("myusername", "mypassword",
                        name="Steve",
                        description="I like cake",
                        icon="person",
                        users_read=True,
                    )

        Args:
            username (str): The username of the new user. Required.
            password (str): The password of the new user. Required.
            name (str,""): The display name for the new user.
            description (str,""): A description for the user.
            icon (str,""): The icon to use for the user.
            public_read (bool,False): Whether the user can be seen by anyone visiting this heedy instance
            users_read (bool,False): Whether the user can be seen by other users
        Returns:
            The :class:`~heedy.User` object for the new user, with all its data cached.
        Throws:
            HeedyException: If insufficient permissions or the request fails.
        """
        return self._create(
            f=lambda x: User(x["username"], session=self.session, cached_data=x),
            **{"username": username, "password": password, **kwargs},
        )
