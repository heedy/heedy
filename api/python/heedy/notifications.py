from .base import Session, APIObject, APIList, q
from typing import Dict


class Notification(APIObject):
    props = {"key"}
    read_qparams = {}

    def __init__(self, cached_data: Dict, session: Session):
        constraints = {"key": cached_data["key"]}
        if "object" in cached_data:
            constraints["object"] = cached_data["object"]
        elif "app" in cached_data:
            constraints["app"] = cached_data["app"]
        else:
            constraints["user"] = cached_data["user"]
        super().__init__(
            "/api/notifications", constraints, session, cached_data=cached_data
        )


class Notifications(APIList):
    """
    Heedy supports notifications for users, apps and objects.
    Notifications can be limited to the related app/object, or can
    be displayed globally in the notification area.

    Notifications are also the main way plugins interact with Heedy's frontend,
    allowing buttons that call the backend or display forms that the user can fill in.

    """

    def __init__(self, constraints: Dict, session: Session):
        super().__init__("/api/notifications", constraints, session)

    def __call__(self, **kwargs):
        """
        Returns a list of notifications satisfying the given constraints.
        All arguments are optional, and constrain the returned results.

        .. tab:: Sync

            ::

                notifications = app.notifications(seen=False,type="md")

        .. tab:: Async

            ::

                notifications = await app.notifications(seen=False,type="md")

        When querying notifications for a specific user or app, all notifications
        of apps and objects owned by the user/app can be returned in a single query
        by specifying the ``object`` or ``app`` constraint to be `*`.

        .. tab:: Sync

            ::

                # Get notifications for all objects belonging to the app
                notifications = app.notifications(objects="*")
                # Get all notifications for the objects belonging to the app,
                # AND notifications for the app itself
                notifications = app.notifications(objects="*",include_self=True)

        .. tab:: Async

            ::

                # Get notifications for all objects belonging to the app
                notifications = await app.notifications(objects="*")
                # Get all notifications for the objects belonging to the app,
                # AND notifications for the app itself
                notifications = await app.notifications(objects="*",include_self=True)

        Args:
            key (str): The key of the notifications to return.
            user (str): Return notifications belonging to the given user.
            app (str): Return notifications belonging to the given app.
            object (str): Return notifications belonging to the given object.
            global (bool): Return notifications that are global or not
            seen (bool): Return only seen/unseen notifications
            dismissible (bool): Return only dismissible/undismissible notifications
            type (str): Return only notifications of the given type (link/md/post)
            include_self (bool): Whether to include the notifications of a constrained user/app when using '*'.
        Returns:
            A list containing the data of all matching notifications.
        Throws:
            HeedyException: If the server returns an error.
        """
        return self._call(
            kwargs, f=lambda x: [Notification(n, self.session) for n in x]
        )

    def __getitem__(self, key: str):
        """
        Returns a notification by its key. This is to be used only when constrained
        to a single user/app/object's notifications, since keys are not globally unique.

        .. tab:: Sync

            ::

                try:
                    notification = app.notifications["my_notification"]
                except KeyError:
                    print("No such notification")

        .. tab:: Async

            ::

                try:
                    notification = await app.notifications["my_notification"]
                except KeyError:
                    print("No such notification")
        """

        def rke(k: str):
            raise KeyError(k)

        return self.session.f(self(key=key), lambda x: x[0] if len(x) > 0 else rke(key))

    def notify(self, key: str, title: str = None, **kwargs):
        """
        Equivalent to ``create(kwargs,overwrite=True)``
        """
        data = {"key": key, **kwargs}
        if title is not None:
            data["title"] = title
        return self.create(data, overwrite=True)

    def create(self, notification, overwrite=False):
        """
        Creates a new notification. Returns an error if the notification already exists.
        If `overwrite` is True, then replaces the existing notification.
        """
        if not "key" in notification:
            raise ValueError("Notification must have a key")
        return self._create(
            notification,
            params={"overwrite": overwrite},
            f=lambda x: Notification(x, self.session),
        )

    def update(self, data, key: str = None, **kwargs):
        """
        Modifies all notifications satisfying the given constraints.
        """
        if key is not None:
            kwargs["key"] = key
        return self.session.patch(
            "api/notifications", data, {**self._constraints, **kwargs}
        )

    def delete(self, key: str = None, **kwargs):
        """
        Deletes notifications satisfying the given constraints.
        Most common usage is deleting a specific notification for a user/app/object identified by its key:

        .. tab:: Sync

            ::

                app.notifications.delete("my_notification")

        .. tab:: Async

            ::

                await app.notifications.delete("my_notification")

        The delete method has identical arguments as the ``__call__`` method, so given a set of constraints,
        it removes all the notifications that would be returned by ``__call__``.

        Raises:
            HeedyException: If the server returns an error.
        """
        if key is not None:
            kwargs["key"] = key
        return self.session.delete(
            "/api/notifications", params={**self.constraints, **kwargs}
        )
