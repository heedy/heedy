from .base import Session
from typing import Dict


class Notifications:
    """
    Heedy supports notifications for users, apps and objects.
    Notifications can be limited to the the related app/object, or can
    be displayed globally in the notification area.

    Notifications are also the main way plugins interact with Heedy's frontend,
    allowing buttons that call the backend or display forms that the user can fill in.

    """
    def __init__(self, constraints: Dict, session: Session):
        self.constraints = constraints
        self.session = session

    def __call__(self, **kwargs):
        """
        Returns a list of notifications satisfying the given constraints.
        """
        return self.session.get(
            "api/notifications", params={**self.constraints, **kwargs}
        )

    def __getitem__(self, key: str):
        """
        Returns a notification by its key. This is to be used only when constrained
        to a single user's/app's or object's notifications, since keys are not globally unique.
        """
        def rke(k):
            raise KeyError(k)
        return self.session.f(self(key=key),lambda x: x[0] if len(x) > 0 else rke(key))

    def notify(self, key: str, title: str = None, **kwargs):
        """
        Alternative to ``create``.
        """
        return self.create(key, title, **kwargs)

    def create(self,key: str,title: str = None, **kwargs):
        """
        Creates a notification
        """
        n = {**self.constraints, **{"key": key}, **kwargs}
        if title is not None:
            n["title"] = title
        if "_global" in n:
            n["global"] = n["_global"]
            del n["_global"]
        return self.session.post("api/notifications", n)

    def update(self,data, **kwargs):
        """
        Modifies the notification
        """
        n = {**self.constraints, **kwargs}
        if "_global" in n:
            n["global"] = n["_global"]
            del n["_global"]
        return self.session.patch("api/notifications",data, n)

    def delete(self, key: str = None, **kwargs):
        """
        Deletes notifications satisfying the given constraints.
        Most common usage is deleting a specific notification identified by its key:

        .. tab:: Sync

            ::

                app.notifications.delete("my_notification")

        .. tab:: Async

            ::

                await app.notifications.delete("my_notification")

        The delete method has identical semantics to the ``__call__`` method, so given a set of constraints,
        it removes the notifications that would be returned by ``__call__``.

        Raises:
            HeedyException: If the server returns an error.
        """
        if key is not None:
            kwargs["key"] = key
        return self.session.delete(
            "/api/notifications", params={**self.constraints, **kwargs}
        )
