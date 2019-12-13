from .base import Session
from typing import Dict

class Notifications:
    def __init__(self,constraints : Dict, session : Session):
        self.constraints = constraints
        self.session = session
    def __call__(self,**kwargs):
        return self.session.get("api/heedy/v1/notifications",params={**self.constraints,**kwargs})
    def notify(self,key,title=None,**kwargs):
        n = {
            **self.constraints,
            **{'key': key},
            **kwargs
        }
        if title is not None:
            n["title"] = title
        if "_global" in n:
            n["global"] = n["_global"]
            del n["_global"]
        return self.session.post("api/heedy/v1/notifications",n)
    def delete(self,key=None,**kwargs):
        if key is not None:
            kwargs["key"] = key
        return self.session.delete("/api/heedy/v1/notifications", params={**self.constraints,**kwargs})