from .base import APIObject


class User(APIObject):
    def __init__(self, session, username):
        super().__init__(session, "api/heedy/v1/users/{username}")

        # Apps represents a list of the user's active apps. Apps can be accessed by ID::
        #   myapp = await u.apps["appid"]
        # They can also by queried, which will return a list::
        #   myapp = await u.apps(plugin="myplugin")
        # Finally, they can be created::
        #   myapp = await u.apps.create()
        self.apps = None
