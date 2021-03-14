from ..base import APIObject, Session, q
from .objects import Object
from .registry import registerObjectType


class DashboardElement(APIObject):
    props = {"query", "settings", "data", "id", "type", "title", "on_demand", "index"}

    def __init__(self, dashboard, cached_data={}):

        super().__init__(
            f"api/objects/{q(dashboard.id)}/dashboard/{q(cached_data['id'])}",
            {},
            dashboard.session,
            cached_data,
        )

        # Dashboard elements don't have notifications
        del self.notifications


class Dashboard(Object):
    def __getitem__(self, getrange):
        """Allows accessing the dashboard's elements as if it were a big python array"""
        if isinstance(getrange, str):
            # The item is an ID, so return the dashboard element
            return self.session.get(
                self.uri + "/dashboard/" + q(getrange),
                f=lambda x: DashboardElement(self, x),
            )

        return self.session.get(
            self.uri + "/dashboard",
            f=lambda x: [DashboardElement(self, xx) for xx in x][getrange],
        )

    def __setitem__(self, elementid, value):
        return self.session.post(self.uri + "/dashboard/" + q(elementid), value)

    def add(self, query, etype="dataset", **kwargs):
        kwargs["type"] = etype
        kwargs["query"] = query
        return self.session.post(self.uri + "/dashboard", [kwargs])


registerObjectType("dashboard", Dashboard)
